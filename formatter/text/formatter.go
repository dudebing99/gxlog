package text

import (
	"regexp"
	"strings"

	"github.com/gratonos/gxlog"
)

const DefaultHeader = "{{time}} {{level}} {{pathname}}:{{line}} {{func}} {{msg}}\n"

var gHeaderRegexp = regexp.MustCompile("{{([^:%]*?)(?::([^%]*?))?(%.*?)?}}")

type Formatter struct {
	headerAppenders []*headerAppender
	prefix          []byte
	suffix          []byte
	buf             []byte
	colorMgr        *colorMgr
	enableColor     bool
}

func New(header string) *Formatter {
	formatter := &Formatter{
		colorMgr: newColorMgr(),
	}
	formatter.SetHeader(header)
	return formatter
}

func (this *Formatter) SetHeader(header string) {
	this.headerAppenders = this.headerAppenders[:0]
	this.prefix = this.prefix[:0]
	this.suffix = this.suffix[:0]
	for header != "" {
		indexes := gHeaderRegexp.FindStringSubmatchIndex(header)
		if indexes == nil {
			this.suffix = append(this.prefix, header...)
			break
		}
		begin, end := indexes[0], indexes[1]
		prefix := header[:begin]
		element, property, fmtspec := extractElement(indexes, header)
		this.addAppender(element, property, fmtspec, prefix)
		header = header[end:]
	}
}

func (this *Formatter) EnableColor() {
	this.enableColor = true
}

func (this *Formatter) DisableColor() {
	this.enableColor = false
}

func (this *Formatter) GetColor(level gxlog.LogLevel) ColorID {
	return this.colorMgr.GetColor(level)
}

func (this *Formatter) SetColor(level gxlog.LogLevel, color ColorID) {
	this.colorMgr.SetColor(level, color)
}

func (this *Formatter) MapColors(colors map[gxlog.LogLevel]ColorID) {
	this.colorMgr.MapColors(colors)
}

func (this *Formatter) Format(record *gxlog.Record) []byte {
	var left, right []byte
	if this.enableColor {
		left, right = this.colorMgr.GetColorEars(record.Level)
	}
	this.buf = this.buf[:0]
	this.buf = append(this.buf, left...)
	for _, appender := range this.headerAppenders {
		this.buf = appender.AppendHeader(this.buf, record)
	}
	this.buf = append(this.buf, this.suffix...)
	this.buf = append(this.buf, right...)
	return this.buf
}

func (this *Formatter) addAppender(element, property, fmtspec, prefix string) {
	this.prefix = append(this.prefix, prefix...)
	appender := newHeaderAppender(element, property, fmtspec, prefix)
	if appender != nil {
		this.headerAppenders = append(this.headerAppenders, appender)
		this.prefix = this.prefix[:0]
	}
}

func extractElement(indexes []int, header string) (element, property, fmtspec string) {
	element = strings.ToLower(getField(header, indexes[2], indexes[3]))
	property = getField(header, indexes[4], indexes[5])
	fmtspec = getField(header, indexes[6], indexes[7])
	if fmtspec == "%" {
		fmtspec = ""
	}
	return element, property, fmtspec
}

func getField(str string, begin, end int) string {
	if begin < end {
		return strings.TrimSpace(str[begin:end])
	}
	return ""
}
