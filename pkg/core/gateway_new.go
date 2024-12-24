package core

type OpenAPISpec struct {
	OpenAPI string              `json:"openapi"           yaml:"openapi"`
	Info    OpenAPIInfo         `json:"info"              yaml:"info"`
	Servers []OpenAPIServer     `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths   map[string]PathItem `json:"paths"             yaml:"paths"` // API paths and operations - translates to a EndpointFile
	// Components OpenAPIComponents   `json:"components,omitempty" yaml:"components,omitempty"`     // not supported
	// Extensions map[string]any      `json:"x-extensions,omitempty" yaml:"x-extensions,omitempty"` // not supported
}

type OpenAPIInfo struct {
	Title       string `json:"title"                 yaml:"title"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Version     string `json:"version"               yaml:"version"`
}

type OpenAPIServer struct {
	URL         string `json:"url"                   yaml:"url"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

type PathItem struct {
	Delete     *Operation     `json:"delete,omitempty"       yaml:"delete,omitempty"`
	Options    *Operation     `json:"options,omitempty"      yaml:"options,omitempty"`
	Put        *Operation     `json:"put,omitempty"          yaml:"put,omitempty"`
	Trace      *Operation     `json:"trace,omitempty"        yaml:"trace,omitempty"`
	Patch      *Operation     `json:"patch,omitempty"        yaml:"patch,omitempty"`
	Get        *Operation     `json:"get,omitempty"          yaml:"get,omitempty"`
	Post       *Operation     `json:"post,omitempty"         yaml:"post,omitempty"`
	Head       *Operation     `json:"head,omitempty"         yaml:"head,omitempty"`
	Connect    *Operation     `json:"connect,omitempty"      yaml:"connect,omitempty"`
	Extensions map[string]any `json:"x-extensions,omitempty" yaml:"x-extensions,omitempty"` // only for direktiv compatibility marker + AllowAnonymous marker  // (Auth + Inbound + Outbound + Target Plugins)
}

type Operation struct {
	Summary     string              `json:"summary,omitempty"     yaml:"summary,omitempty"`
	Description string              `json:"description,omitempty" yaml:"description,omitempty"`
	Parameters  []Parameter         `json:"parameters,omitempty"  yaml:"parameters,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"             yaml:"responses"`
	// Extensions  map[string]any      `json:"x-extensions,omitempty" yaml:"x-extensions,omitempty"` not supported
}

type Parameter struct {
	Name        string `json:"name"                  yaml:"name"`
	In          string `json:"in"                    yaml:"in"`
	Required    bool   `json:"required"              yaml:"required"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Schema      Schema `json:"schema"                yaml:"schema"`
	Example     any    `json:"example,omitempty"     yaml:"example,omitempty"`
}

type RequestBody struct {
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]Schema `json:"content"               yaml:"content"`
	Required    bool              `json:"required"              yaml:"required"`
	// Extensions  []PluginConfig    `json:"x-extensions,omitempty" yaml:"x-extensions,omitempty"` // not supported
}

type Response struct {
	Description string            `json:"description"       yaml:"description"`
	Content     map[string]Schema `json:"content,omitempty" yaml:"content,omitempty"`
	// Extensions  []PluginConfig    `json:"x-extensions,omitempty" yaml:"x-extensions,omitempty"` // not supported
}

type Schema struct {
	Type       string            `json:"type,omitempty"       yaml:"type,omitempty"`
	Properties map[string]Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Items      *Schema           `json:"items,omitempty"      yaml:"items,omitempty"`
	Ref        string            `json:"$ref,omitempty"       yaml:"$ref,omitempty"`
	Example    any               `json:"example,omitempty"    yaml:"example,omitempty"`
}

// type OpenAPIComponents struct {
// 	SecuritySchemes map[string]any    `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"` // Auth schemes
// 	Schemas         map[string]Schema `json:"schemas,omitempty"         yaml:"schemas,omitempty"`         // Reusable schemas
// }

// type SecurityScheme struct {
// 	Type    string         `json:"type"              yaml:"type"`              // Security scheme type ("http", "apiKey", "oauth2", "custom")
// 	Scheme  string         `json:"scheme,omitempty"  yaml:"scheme,omitempty"`  // For HTTP, e.g., "basic", "bearer"
// 	Plugins []PluginConfig `json:"plugins,omitempty" yaml:"plugins,omitempty"` // Auth plugin configuration
// }
