import type { EditorProps } from "@monaco-editor/react";
import { FC } from "react";
import MonacoEditor from "@monaco-editor/react";
import { useTheme } from "../../../../../util/store/theme";

function setEditorTheme(monaco: any) {
  monaco.editor.defineTheme("direktiv-dark", {
    base: "vs-dark",
    inherit: true,
    rules: [],
    colors: {
      "editor.background": "#000000",
    },
  });
}

const Editor: FC<EditorProps> = ({ ...props }) => {
  const theme = useTheme();
  return (
    <MonacoEditor
      beforeMount={setEditorTheme}
      options={{
        scrollBeyondLastLine: false,
        cursorBlinking: "smooth",
        wordWrap: true,
        fontSize: "13px",
        minimap: {
          enabled: false,
        },
      }}
      loading=""
      language="yaml"
      theme={theme === "dark" ? "direktiv-dark" : "vs-light"}
      {...props}
    />
  );
};

export default Editor;
