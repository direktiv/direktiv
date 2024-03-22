import * as monaco from "monaco-editor";

// for reference:
// https://github.com/Microsoft/vscode/blob/913e891c34f8b4fe2c0767ec9f8bfd3b9dbe30d9/src/vs/editor/standalone/common/themes.ts#L13
const theme: monaco.editor.IStandaloneThemeData = {
  base: "vs-dark",
  inherit: true,
  rules: [
    {
      foreground: "505050", // grayDark.gray8
      fontStyle: "italic",
      token: "comment",
    },
    {
      foreground: "6473FF", // primary.400
      token: "number",
    },
    {
      foreground: "6473FF", // primary.400
      token: "string.yaml",
    },
    {
      foreground: "a0a0a0", // grayDark.gray11
      token: "type",
    },
    {
      foreground: "a0a0a0", // grayDark.gray11
      token: "string.key.json", // JSON Key
    },
    {
      foreground: "6473FF", // primary.400
      token: "string.value.json", // JSON Value
    },
    {
      foreground: "a0a0a0", // grayDark.gray11
      token: "tag", // HTML Tag name
    },
    {
      foreground: "a0a0a0", // gray.gray11
      token: "delimiter.html", // HTML Tag <>
    },
    {
      foreground: "a0a0a0", // grayDark.gray11
      token: "metatag.html", // HTML Meta tag
    },
    {
      foreground: "a0a0a0", // grayDark.gray11
      token: "metatag.content.html", // HTML Meta tag content
    },
    {
      foreground: "6473FF", // primary.400
      token: "delimiter", // HTML Meta tag content
    },
    {
      foreground: "a0a0a0", // grayDark.gray11
      token: "attribute.name", // HTML Attribute Name
    },
    {
      foreground: "6473FF", // primary.400
      token: "attribute.value.html", // HTML Attribute Name
    },
    {
      foreground: "30A46C", // green9 - success-dark-9
      token: "comment",
    },
    {
      foreground: "30A46C", // green9 - success-dark-9
      token: "attribute.value",
    },
    {
      foreground: "6473FF", // primary.400
      token: "attribute.value.number", // html attribute value number, e.g. [5]px
    },
    {
      foreground: "6473FF", // primary.400
      token: "attribute.value.unit", // html attribute value unit, e.g. 5[px]
    },
    {
      foreground: "6473FF", // primary.400
      token: "string", // css string value: e.g. font-family: "Segoe UI","HelveticaNeue-Light",
    },
    {
      foreground: "6473FF", // primary.400
      token: "metatag", // metatag in Shell script e.g. #!/bin/bash
    },
    {
      foreground: "6473FF", // primary.400
      token: "keyword", // keyword in Shell script
    },
    {
      foreground: "a0a0a0", // grayDark.gray11
      token: "variable.predefined", // variable defined in Shell script
    },
    {
      foreground: "6473FF", // primary.400
      token: "variable", // Shell script variable
    },
    {
      foreground: "6473FF", // primary.400
      token: "attribute.value.number.css",
    },
    {
      foreground: "6473FF", // primary.400
      token: "attribute.value.unit.css",
    },
    {
      foreground: "6473FF", // primary.400
      token: "attribute.value.hex.css",
    },
  ],
  colors: {
    "editor.foreground": "#a0a0a0", // added for HTML Tag Content gray-dark-10
    "editor.background": "#000000",
    "editor.selectionBackground": "#ffffff2e", // whiteA.whiteA7
  },
};

export default theme;
