package extensions

// AdditionalSchema for hooking additional sql schema provisioning scripts. This helps build new plugins and
// extensions for Direktiv.
var AdditionalSchema func() string
