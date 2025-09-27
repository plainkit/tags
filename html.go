package tags

type AttributeRef struct {
	Name    string
	Boolean bool
}

type HTMLElement struct {
	Empty      bool
	Attributes []AttributeRef
}

type HTMLMeta struct {
	Source         string
	SchemaVersion  string
	ElementCount   int
	AttributeCount int
}

type HTMLIndex struct {
	Meta     HTMLMeta
	Globals  []AttributeRef
	Elements map[string]HTMLElement
}

var HTML = HTMLIndex{
	Meta: HTMLMeta{
		Source:         "https://html.spec.whatwg.org/multipage/indices.html",
		SchemaVersion:  "1.1.0",
		ElementCount:   115,
		AttributeCount: 140,
	},
	Globals: []AttributeRef{
		{Name: "accesskey"},
		{Name: "autocapitalize"},
		{Name: "autocorrect"},
		{Name: "autofocus", Boolean: true},
		{Name: "class"},
		{Name: "contenteditable"},
		{Name: "dir"},
		{Name: "draggable"},
		{Name: "enterkeyhint"},
		{Name: "hidden"},
		{Name: "id"},
		{Name: "inert", Boolean: true},
		{Name: "inputmode"},
		{Name: "is"},
		{Name: "itemid"},
		{Name: "itemprop"},
		{Name: "itemref"},
		{Name: "itemscope", Boolean: true},
		{Name: "itemtype"},
		{Name: "lang"},
		{Name: "nonce"},
		{Name: "popover"},
		{Name: "slot"},
		{Name: "spellcheck"},
		{Name: "style"},
		{Name: "tabindex"},
		{Name: "title"},
		{Name: "translate"},
		{Name: "writingsuggestions"},
	},
	Elements: map[string]HTMLElement{
		"a": {
			Attributes: []AttributeRef{
				{Name: "download"},
				{Name: "href"},
				{Name: "hreflang"},
				{Name: "ping"},
				{Name: "referrerpolicy"},
				{Name: "rel"},
				{Name: "target"},
				{Name: "type"},
			},
		},
		"abbr":    {},
		"address": {},
		"area": {
			Empty: true,
			Attributes: []AttributeRef{
				{Name: "alt"},
				{Name: "coords"},
				{Name: "download"},
				{Name: "href"},
				{Name: "ping"},
				{Name: "referrerpolicy"},
				{Name: "rel"},
				{Name: "shape"},
				{Name: "target"},
			},
		},
		"article": {},
		"aside":   {},
		"audio": {
			Attributes: []AttributeRef{
				{Name: "autoplay", Boolean: true},
				{Name: "controls", Boolean: true},
				{Name: "crossorigin"},
				{Name: "loop", Boolean: true},
				{Name: "muted", Boolean: true},
				{Name: "preload"},
				{Name: "src"},
			},
		},
		"b": {},
		"base": {
			Empty: true,
			Attributes: []AttributeRef{
				{Name: "href"},
				{Name: "target"},
			},
		},
		"bdi": {},
		"bdo": {},
		"blockquote": {
			Attributes: []AttributeRef{
				{Name: "cite"},
			},
		},
		"body": {},
		"br": {
			Empty: true,
		},
		"button": {
			Attributes: []AttributeRef{
				{Name: "command"},
				{Name: "commandfor"},
				{Name: "formaction"},
				{Name: "formenctype"},
				{Name: "formmethod"},
				{Name: "formnovalidate", Boolean: true},
				{Name: "formtarget"},
				{Name: "popovertarget"},
				{Name: "popovertargetaction"},
				{Name: "type"},
				{Name: "value"},
			},
		},
		"canvas": {
			Attributes: []AttributeRef{
				{Name: "height"},
				{Name: "width"},
			},
		},
		"caption": {},
		"cite":    {},
		"code":    {},
		"col": {
			Empty: true,
			Attributes: []AttributeRef{
				{Name: "span"},
			},
		},
		"colgroup": {
			Attributes: []AttributeRef{
				{Name: "span"},
			},
		},
		"data": {
			Attributes: []AttributeRef{
				{Name: "value"},
			},
		},
		"datalist": {},
		"dd":       {},
		"del": {
			Attributes: []AttributeRef{
				{Name: "cite"},
				{Name: "datetime"},
			},
		},
		"details": {
			Attributes: []AttributeRef{
				{Name: "name"},
				{Name: "open", Boolean: true},
			},
		},
		"dfn": {},
		"dialog": {
			Attributes: []AttributeRef{
				{Name: "closedby"},
				{Name: "open", Boolean: true},
			},
		},
		"div": {},
		"dl":  {},
		"dt":  {},
		"em":  {},
		"embed": {
			Empty: true,
			Attributes: []AttributeRef{
				{Name: "height"},
				{Name: "src"},
				{Name: "type"},
				{Name: "width"},
			},
		},
		"fieldset": {
			Attributes: []AttributeRef{
				{Name: "disabled", Boolean: true},
			},
		},
		"figcaption": {},
		"figure":     {},
		"footer":     {},
		"form": {
			Attributes: []AttributeRef{
				{Name: "accept-charset"},
				{Name: "action"},
				{Name: "autocomplete"},
				{Name: "enctype"},
				{Name: "method"},
				{Name: "name"},
				{Name: "novalidate", Boolean: true},
				{Name: "target"},
			},
		},
		"h1":     {},
		"h2":     {},
		"h3":     {},
		"h4":     {},
		"h5":     {},
		"h6":     {},
		"head":   {},
		"header": {},
		"hgroup": {},
		"hr": {
			Empty: true,
		},
		"html": {},
		"i":    {},
		"iframe": {
			Empty: true,
			Attributes: []AttributeRef{
				{Name: "allow"},
				{Name: "allowfullscreen", Boolean: true},
				{Name: "height"},
				{Name: "loading"},
				{Name: "name"},
				{Name: "referrerpolicy"},
				{Name: "sandbox"},
				{Name: "src"},
				{Name: "srcdoc"},
				{Name: "width"},
			},
		},
		"img": {
			Empty: true,
			Attributes: []AttributeRef{
				{Name: "alt"},
				{Name: "crossorigin"},
				{Name: "decoding"},
				{Name: "fetchpriority"},
				{Name: "height"},
				{Name: "ismap", Boolean: true},
				{Name: "loading"},
				{Name: "referrerpolicy"},
				{Name: "sizes"},
				{Name: "src"},
				{Name: "srcset"},
				{Name: "usemap"},
				{Name: "width"},
			},
		},
		"input": {
			Empty: true,
			Attributes: []AttributeRef{
				{Name: "accept"},
				{Name: "alpha", Boolean: true},
				{Name: "alt"},
				{Name: "autocomplete"},
				{Name: "checked", Boolean: true},
				{Name: "colorspace"},
				{Name: "dirname"},
				{Name: "formaction"},
				{Name: "formenctype"},
				{Name: "formmethod"},
				{Name: "formnovalidate", Boolean: true},
				{Name: "formtarget"},
				{Name: "height"},
				{Name: "list"},
				{Name: "max"},
				{Name: "maxlength"},
				{Name: "min"},
				{Name: "minlength"},
				{Name: "multiple", Boolean: true},
				{Name: "pattern"},
				{Name: "placeholder"},
				{Name: "popovertarget"},
				{Name: "popovertargetaction"},
				{Name: "readonly", Boolean: true},
				{Name: "required", Boolean: true},
				{Name: "size"},
				{Name: "src"},
				{Name: "step"},
				{Name: "type"},
				{Name: "value"},
				{Name: "width"},
			},
		},
		"ins": {
			Attributes: []AttributeRef{
				{Name: "cite"},
				{Name: "datetime"},
			},
		},
		"kbd": {},
		"label": {
			Attributes: []AttributeRef{
				{Name: "for"},
			},
		},
		"legend": {},
		"li": {
			Attributes: []AttributeRef{
				{Name: "value"},
			},
		},
		"link": {
			Empty: true,
			Attributes: []AttributeRef{
				{Name: "as"},
				{Name: "blocking"},
				{Name: "color"},
				{Name: "crossorigin"},
				{Name: "disabled", Boolean: true},
				{Name: "fetchpriority"},
				{Name: "href"},
				{Name: "hreflang"},
				{Name: "imagesizes"},
				{Name: "imagesrcset"},
				{Name: "integrity"},
				{Name: "media"},
				{Name: "referrerpolicy"},
				{Name: "rel"},
				{Name: "sizes"},
				{Name: "type"},
			},
		},
		"main": {},
		"map": {
			Attributes: []AttributeRef{
				{Name: "name"},
			},
		},
		"mark": {},
		"math": {},
		"menu": {},
		"meta": {
			Empty: true,
			Attributes: []AttributeRef{
				{Name: "charset"},
				{Name: "content"},
				{Name: "http-equiv"},
				{Name: "media"},
				{Name: "name"},
			},
		},
		"meter": {
			Attributes: []AttributeRef{
				{Name: "high"},
				{Name: "low"},
				{Name: "max"},
				{Name: "min"},
				{Name: "optimum"},
				{Name: "value"},
			},
		},
		"nav":      {},
		"noscript": {},
		"object": {
			Attributes: []AttributeRef{
				{Name: "data"},
				{Name: "height"},
				{Name: "name"},
				{Name: "type"},
				{Name: "width"},
			},
		},
		"ol": {
			Attributes: []AttributeRef{
				{Name: "reversed", Boolean: true},
				{Name: "start"},
				{Name: "type"},
			},
		},
		"optgroup": {
			Attributes: []AttributeRef{
				{Name: "label"},
			},
		},
		"option": {
			Attributes: []AttributeRef{
				{Name: "label"},
				{Name: "selected", Boolean: true},
				{Name: "value"},
			},
		},
		"output": {
			Attributes: []AttributeRef{
				{Name: "for"},
			},
		},
		"p":       {},
		"picture": {},
		"pre":     {},
		"progress": {
			Attributes: []AttributeRef{
				{Name: "max"},
				{Name: "value"},
			},
		},
		"q": {
			Attributes: []AttributeRef{
				{Name: "cite"},
			},
		},
		"rp":   {},
		"rt":   {},
		"ruby": {},
		"s":    {},
		"samp": {},
		"script": {
			Attributes: []AttributeRef{
				{Name: "async", Boolean: true},
				{Name: "blocking"},
				{Name: "crossorigin"},
				{Name: "defer", Boolean: true},
				{Name: "fetchpriority"},
				{Name: "integrity"},
				{Name: "nomodule", Boolean: true},
				{Name: "referrerpolicy"},
				{Name: "src"},
				{Name: "type"},
			},
		},
		"search":  {},
		"section": {},
		"select": {
			Attributes: []AttributeRef{
				{Name: "autocomplete"},
				{Name: "multiple", Boolean: true},
				{Name: "required", Boolean: true},
				{Name: "size"},
			},
		},
		"selectedcontent": {
			Empty: true,
		},
		"slot": {
			Attributes: []AttributeRef{
				{Name: "name"},
			},
		},
		"small": {},
		"source": {
			Empty: true,
			Attributes: []AttributeRef{
				{Name: "height"},
				{Name: "media"},
				{Name: "sizes"},
				{Name: "src"},
				{Name: "srcset"},
				{Name: "type"},
				{Name: "width"},
			},
		},
		"span":   {},
		"strong": {},
		"style": {
			Attributes: []AttributeRef{
				{Name: "blocking"},
				{Name: "media"},
			},
		},
		"sub":     {},
		"summary": {},
		"sup":     {},
		"svg":     {},
		"table":   {},
		"tbody":   {},
		"td": {
			Attributes: []AttributeRef{
				{Name: "colspan"},
				{Name: "headers"},
				{Name: "rowspan"},
			},
		},
		"template": {
			Empty: true,
			Attributes: []AttributeRef{
				{Name: "shadowrootclonable", Boolean: true},
				{Name: "shadowrootcustomelementregistry", Boolean: true},
				{Name: "shadowrootdelegatesfocus", Boolean: true},
				{Name: "shadowrootmode"},
				{Name: "shadowrootserializable", Boolean: true},
			},
		},
		"textarea": {
			Attributes: []AttributeRef{
				{Name: "autocomplete"},
				{Name: "cols"},
				{Name: "dirname"},
				{Name: "maxlength"},
				{Name: "minlength"},
				{Name: "placeholder"},
				{Name: "readonly", Boolean: true},
				{Name: "required", Boolean: true},
				{Name: "rows"},
				{Name: "wrap"},
			},
		},
		"tfoot": {},
		"th": {
			Attributes: []AttributeRef{
				{Name: "abbr"},
				{Name: "colspan"},
				{Name: "headers"},
				{Name: "rowspan"},
				{Name: "scope"},
			},
		},
		"thead": {},
		"time": {
			Attributes: []AttributeRef{
				{Name: "datetime"},
			},
		},
		"title": {},
		"tr":    {},
		"track": {
			Empty: true,
			Attributes: []AttributeRef{
				{Name: "default", Boolean: true},
				{Name: "kind"},
				{Name: "label"},
				{Name: "src"},
				{Name: "srclang"},
			},
		},
		"u":   {},
		"ul":  {},
		"var": {},
		"video": {
			Attributes: []AttributeRef{
				{Name: "autoplay", Boolean: true},
				{Name: "controls", Boolean: true},
				{Name: "crossorigin"},
				{Name: "height"},
				{Name: "loop", Boolean: true},
				{Name: "muted", Boolean: true},
				{Name: "playsinline", Boolean: true},
				{Name: "poster"},
				{Name: "preload"},
				{Name: "src"},
				{Name: "width"},
			},
		},
		"wbr": {
			Empty: true,
		},
	},
}
