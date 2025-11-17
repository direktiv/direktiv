import parserTypescript from "prettier/plugins/typescript";
import pluginEstree from "prettier/plugins/estree";
import prettier from "prettier/standalone";

export const formatTs = (code: string) =>
  prettier.format(code, {
    parser: "typescript",
    plugins: [pluginEstree, parserTypescript],
  });
