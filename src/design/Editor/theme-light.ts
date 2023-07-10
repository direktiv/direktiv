// for reference:
// https://github.com/Microsoft/vscode/blob/913e891c34f8b4fe2c0767ec9f8bfd3b9dbe30d9/src/vs/editor/standalone/common/themes.ts#L13
export default {
  base: "vs",
  inherit: true,
  rules: [
    {
      foreground: "C7C7C7", // gray.gray8
      fontStyle: "italic",
      token: "comment",
    },
    {
      foreground: "5364FF", // primary.500
      token: "number",
    },
    {
      foreground: "5364FF", // primary.500
      token: "string.yaml",
    },
    {
      foreground: "6f6f6f", // gray.gray11
      token: "type",
    },
    {
      foreground: "6f6f6f", // gray.gray11
      token: "string.key.json", // JSON Key
    },
    {
      foreground: "5364FF",
      token: "string.value.json", // JSON Value
    },
    {
      foreground: "6f6f6f", // gray.gray11
      token: "tag", // HTML Tag name
    },
    {
      foreground: "6f6f6f", // gray.gray11
      token: "delimiter.html", // HTML Tag <>
    },
    {
      foreground: "6f6f6f", // gray.gray11
      token: "metatag.html", // HTML Meta tag
    },
    {
      foreground: "6f6f6f", // gray.gray11
      token: "metatag.content.html", // HTML Meta tag content
    },
    {
      foreground: "5364FF", // primary.500
      token: "delimiter", // HTML Meta tag content
    },
    {
      foreground: "6f6f6f", // gray.gray11
      token: "attribute.name", // HTML Attribute Name
    },
    {
      foreground: "5364FF", // primary.500
      token: "attribute.value.html", // HTML Attribute Name
    },
    {
      foreground: "30A46C", // success-9
      token: "comment",
    },
    {
      foreground: "30A46C", // success-9
      token: "attribute.value",
    },
    {
      foreground: "5364FF", // primary.500
      token: "attribute.value.number", // html attribute value number, e.g. [5]px
    },
    {
      foreground: "5364FF", // primary.500
      token: "attribute.value.unit", // html attribute value unit, e.g. 5[px]
    },
    {
      foreground: "5364FF", // primary.500
      token: "string", // css string value: e.g. font-family: "Segoe UI","HelveticaNeue-Light",
    },
    {
      foreground: "5364FF", // primary.500
      token: "metatag", // metatag in Shell script e.g. #!/bin/bash
    },
    {
      foreground: "5364FF", // primary.500
      token: "keyword", // keyword in Shell script
    },
    {
      foreground: "6f6f6f", // gray.gray11
      token: "variable.predefined", // variable defined in Shell script
    },
    {
      foreground: "5364FF", // primary.500
      token: "variable", // Shell script variable
    },
    {
      foreground: "5364FF", // primary.500
      token: "attribute.value.number.css",
    },
    {
      foreground: "5364FF", // primary.500
      token: "attribute.value.unit.css",
    },
    {
      foreground: "5364FF", // primary.500
      token: "attribute.value.hex.css",
    },
  ],
  colors: {
    "editor.foreground": "#6f6f6f", // added for HTML Tag Content
    "editor.background": "#ffffff",
    "editor.selectionBackground": "#00000012", // blackA.blackA4
  },
} as const;
