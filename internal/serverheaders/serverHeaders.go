package serverheaders

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type ServerHandler struct {
	metricMapGauge   map[string]float64
	metricMapCounter map[string]int64
}

func NewServerHandler() *ServerHandler {
	p := new(ServerHandler)
	p.metricMapCounter = make(map[string]int64)
	p.metricMapGauge = make(map[string]float64)
	return p
}

func (h *ServerHandler) NotFoundHandler(c echo.Context) error {
	return c.NoContent(http.StatusNotFound)
}

func (h *ServerHandler) MainHandler(c echo.Context) error {

	tableHead := "<table align=\"center\" border=\"1\" cellpadding=\"1\" cellspacing=\"1\" style=\"width:100%\">\n<thead>\n<tr>\n<th scope=\"col\">ID</th>\n<th scope=\"col\">Metric name</th>\n<th scope=\"col\">Value</th>\n<th scope=\"col\">Hyperlinc</th>\n</tr>\n</thead>\n<tbody>\n"
	tableEnd := "</tbody>\n</table>\n"
	result := "<html>\n<body>\n<h1>GAUGES:</h1>\n" + tableHead
	i := 0
	for id, val := range h.metricMapGauge {
		result += fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%f</td><td><a href=\"http://localhost:8080/value/gauge/%s\">http://localhost:8080/value/gauge/%s</a></td></tr>\n",
			i, id, val, id, id)
		i++
	}
	result += tableEnd
	result += "<h1>COUNTERS:</h1>\n"
	result += tableHead
	i = 0
	for id, val := range h.metricMapCounter {
		result += fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%d</td><td><a href=\"http://localhost:8080/value/gauge/%s\">http://localhost:8080/value/gauge/%s</a></td></tr>\n",
			i, id, val, id, id)
		i++
	}
	result += tableEnd
	result += "</html>\n</body>"

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return c.String(http.StatusOK, result)
}

func (h *ServerHandler) GetHandler(c echo.Context) error {
	if c.Param("method") != "value" {
		return c.NoContent(http.StatusNotFound)
	}
	switch typeM := c.Param("type"); typeM {
	case "counter":
		val, ok := h.metricMapCounter[c.Param("metric")]
		if !ok {
			return c.NoContent(http.StatusNotFound)
		}
		return c.String(http.StatusOK, strconv.FormatInt(val, 10))
	case "gauge":
		val, ok := h.metricMapGauge[c.Param("metric")]
		if !ok {
			return c.NoContent(http.StatusNotFound)
		}
		return c.String(http.StatusOK, strconv.FormatFloat(val, 'f', -1, 64))
	default:
		return c.NoContent(http.StatusNotImplemented)
	}
}

func (h *ServerHandler) UpdateHandler(c echo.Context) error {

	if c.Param("method") != "update" {
		return c.NoContent(http.StatusNotFound)
	}
	switch typeM := c.Param("type"); typeM {
	case "counter":
		val, err := strconv.ParseInt(c.Param("value"), 0, 64)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		h.metricMapCounter[c.Param("metric")] += val
		return c.NoContent(http.StatusOK)
	case "gauge":
		val, err := strconv.ParseFloat(c.Param("value"), 64)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		h.metricMapGauge[c.Param("metric")] = val
		return c.NoContent(http.StatusOK)
	default:
		return c.NoContent(http.StatusNotImplemented)
	}
}
