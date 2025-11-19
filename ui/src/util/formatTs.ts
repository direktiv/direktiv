import parserTypescript from "prettier/plugins/typescript";
import pluginEstree from "prettier/plugins/estree";
import prettier from "prettier/standalone";

export const formatTs = (code: string) =>
  prettier.format(code, {
    parser: "typescript",
    plugins: [pluginEstree, parserTypescript],
    // if needed, define overrides for the default settings here,
    // see https://prettier.io/docs/options.
  });
