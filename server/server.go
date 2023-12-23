package server

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"github.com/vooon/esp-ota-server/assets"
)

//go:generate rice embed-go

type server struct {
	config    Config
	templates *template.Template
	registry  *TTLMap
}

// Render renders a template document
func (s server) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	// Add global methods if data is a map
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}

	return s.templates.ExecuteTemplate(w, name, data)
}

func getEspHeader(hdr http.Header, key string) (ret string, ok bool) {
	var val []string
	val, ok = hdr[http.CanonicalHeaderKey("x-esp8266-"+key)]
	if !ok {
		val, ok = hdr[http.CanonicalHeaderKey("x-esp32-"+key)]
		if !ok {
			ret = ""
			return
		}
	}
	ret = val[0]
	return
}

func (s server) register(context echo.Context) error {
	logger := context.Logger()
	buf, err := ioutil.ReadAll(context.Request().Body)
	if err != nil {
		logger.Error("Could not read request body")
		return context.String(http.StatusBadRequest, "Invalid Request")
	}

	jsonMap := make(map[string]interface{})
	err = json.NewDecoder(bytes.NewReader(buf)).Decode(&jsonMap)
	if err != nil {
		logger.Warn("Invalid json: ", string(buf))
		return context.String(http.StatusBadRequest, "Invalid json")
	}
	logger.Print("Valid json")

	hdr := context.Request().Header
	staMac, macOk := getEspHeader(hdr, "sta-mac")
	if !macOk {
		logger.Warn("Missing MAC header")
		return context.String(http.StatusBadRequest, "Missing MAC header")
	}

	ipMap := jsonMap["ip"]
	ip, ok := ipMap.(string)
	if !ok {
		logger.Warn("Missing id in json: ", string(buf))
		return context.String(http.StatusBadRequest, "Missing ip in JSON body")
	}
	networkMap := jsonMap["network"]
	network, ok := networkMap.(string)
	if !ok {
		logger.Warn("Missing network in json: ", string(buf))
		return context.String(http.StatusBadRequest, "Missing network in JSON body")
	}

	logger.Print("Welcome: IP " + ip + " (mac: " + staMac + ") from network " + network)
	s.registry.Put(strings.ToLower(network), staMac, ip)

	return context.String(http.StatusCreated, "Registered")
}

func (s server) lookup(context echo.Context) error {
	logger := context.Logger()
	network := context.Param("network")
	ips := s.registry.Get(strings.ToLower(network))
	if len(ips) == 0 {
		escNetwork, errE := url.QueryUnescape(network)
		if nil == errE {
			network = escNetwork
			ips = s.registry.Get(strings.ToLower(network))
		}
		if len(ips) == 0 {
			logger.Print("Did not find '" + network + "' in " + s.registry.Keys())
			return context.String(http.StatusNotFound, "No ips registered for "+network)
		}
	}
	if len(ips) == 1 {
		return context.Redirect(http.StatusFound, "http://"+ips[0])
	}

	return context.Render(http.StatusOK, "iplist.ghtm", map[string]interface{}{
		"network": network,
		"ips":     ips,
	})
}

func (s server) getBinaryFile(context echo.Context) error {
	logger := context.Logger()

	project := context.Param("project")
	filename := context.Param("file")

	path := filepath.Join(s.config.DataDirPath, "bin", project, filename)
	file, err := os.Open(path)
	if err != nil && os.IsNotExist(err) {
		logger.Warnj(log.JSON{
			"msg":       "File not found",
			"err":       err,
			"file_path": path,
		})
		return context.String(http.StatusNotFound, "no file")
	} else if err != nil {
		return err
	}
	defer file.Close()

	md5hasher := md5.New()

	teeRd := io.TeeReader(file, md5hasher)

	b, err := io.ReadAll(teeRd)
	if err != nil {
		return err
	}

	md5sum := hex.EncodeToString(md5hasher.Sum(nil))

	hdr := context.Request().Header

	logger.Printj(log.JSON{
		"esp_request_headers": hdr,
	})

	staMac, macOk := getEspHeader(hdr, "sta-mac")
	//apMac, _ := hdr["X-Esp8266-Ap-Mac"]
	//freeSpace, _ := hdr["X-Esp8266-Free-Space"]
	//sketchSize, _ := hdr["X-Esp8266-Sketch-Size"]
	sketchMd5, md5ok := getEspHeader(hdr, "sketch-md5")
	//chipSize, _ := hdr["X-Esp8266-Chip-Size"]
	//sdkVersion, _ := hdr["X-Esp8266-Sdk-Version"]

	/**
	 * mode can be one of:
	 * - sketch: download sketch
	 * - spiffs: ask for filesystem image
	 * - version: ask for available version. expects answer in x-version header,
	 */
	mode, modeOk := getEspHeader(hdr, "mode")
	version, versionOk := getEspHeader(hdr, "version")

	if !modeOk {
		return context.String(http.StatusBadRequest, "bad request")
	}
	if "sketch" != mode {
		logger.Info("Mode " + mode + " not implemented")
		return s.get422(context)
	}

	sendFile := true
	if versionOk {
		// TODO version handling
		// also set x-version in response
	} else {
		version = "-unspecified-"

	}
	if md5ok {
		sendFile = sketchMd5 != md5sum
	} else {
		sketchMd5 = "-unspecified-"
	}
	if !macOk {
		staMac = "-unspecified-"
	}

	context.Response().Header()["x-MD5"] = []string{md5sum} // do not do strings.Title()
	logger.Printj(log.JSON{
		"esp_mac":        staMac,
		"esp_mode":       mode,
		"esp_version":    version,
		"esp_sketch_md5": sketchMd5,
		"our_md5":        md5sum,
		"send_file":      sendFile,
		"file_path":      path,
		"file_size":      len(b),
	})

	if sendFile {
		return context.File(path)
	} else {
		return context.String(http.StatusNotModified, "")
	}
}

func (s server) getIndex(context echo.Context) error {
	return context.Render(http.StatusOK, "index.ghtm", map[string]interface{}{})
}
func (s server) get422(context echo.Context) error {
	return context.String(http.StatusUnprocessableEntity, "Can not handle the request")
}

func parseTemplates() (*template.Template, error) {
	return template.ParseFS(assets.Assets, "*.ghtm")
}

func Serve(config Config) error {
	echoServer := echo.New()

	echoServer.Use(middleware.Logger())
	echoServer.Use(middleware.Recover())

	newpath, err := filepath.Abs(config.DataDirPath)
	if err != nil {
		echoServer.Logger.Fatal("can't abs data-dir")
		return err
	}
	if stat, err := os.Stat(newpath); err == nil && stat.IsDir() {
		echoServer.Logger.Info("Data-dir: ", newpath)
		config.DataDirPath = newpath
	} else {
		echoServer.Logger.Fatal("data-dir not exist! ", newpath)
		return err
	}

	templates, err := parseTemplates()
	if err != nil {
		return err
	}

	s := server{
		config:    config,
		templates: templates,
		registry:  CreateTTLMap(3600),
	}

	assetHandler := http.FileServer(http.FS(assets.Assets))

	echoServer.Renderer = s
	echoServer.GET("/bin/:project/:file", s.getBinaryFile)
	echoServer.GET("/assets/*", echo.WrapHandler(http.StripPrefix("/assets/", assetHandler)))
	echoServer.GET("/", s.getIndex)
	echoServer.POST("/register", s.register)
	echoServer.GET("/lookup/:network", s.lookup)

	return echoServer.Start(config.Bind)
}
