import AutoSizer from "react-virtualized-auto-sizer";
import type { EditorProps } from "@monaco-editor/react";
import { FC } from "react";
import MonacoEditor from "@monaco-editor/react";

const beforeMount: EditorProps["beforeMount"] = (monaco) => {
  monaco.editor.defineTheme("direktiv-dark", {
    base: "vs-dark",
    inherit: true,
    rules: [],
    colors: {
      "editor.background": "#000000",
      "editor.selectionBackground": "#ffffff2e", // whiteA.whiteA7
    },
  });
};

const onMount: EditorProps["onMount"] = (editor, monaco) => {
  editor.addCommand(monaco.KeyCode.KEY_S, () => {
    alert("you've the s key");
  });
};

const Editor: FC<
  Omit<EditorProps, "beforeMount" | "onMount"> & { theme?: "light" | "dark" }
> = ({ options, theme, ...props }) => (
  <AutoSizer>
    {({ height, width }) => (
      <MonacoEditor
        width={width}
        height={height}
        beforeMount={beforeMount}
        onMount={onMount}
        options={{
          scrollBeyondLastLine: false,
          cursorBlinking: "smooth",
          wordWrap: true,
          fontSize: "13px",
          minimap: {
            enabled: false,
          },
          ...options,
        }}
        loading=""
        language="yaml"
        theme={theme === "dark" ? "direktiv-dark" : "vs-light"}
        {...props}
      />
    )}
  </AutoSizer>
);

export default Editor;
