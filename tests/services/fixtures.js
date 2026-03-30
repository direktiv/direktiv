export const bashServiceSrc = `{
  "image": "direktiv/bash:dev",
  "scale": 1,
  "size": "small"
}`

export const bashServiceScale3Src = `{
  "image": "direktiv/bash:dev",
  "scale": 2,
  "size": "small"
}`

export const bashServiceScale0Src = `{
  "image": "direktiv/bash:dev",
  "scale": 0,
  "size": "small"
}`

export const echoServiceSrc = `{
  "image": "direktiv/echo",
  "scale": 1,
  "size": "small"
}`

export const bashServiceWithEnvsSrc = `{
  "image": "direktiv/bash:dev",
  "scale": 1,
  "size": "small",
  "envs": [
    { "name": "FOO1", "value": "bar1" },
    { "name": "FOO2", "value": "bar2" }
  ]
}`
