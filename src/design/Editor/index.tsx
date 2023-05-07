import type { EditorProps } from "@monaco-editor/react";
import { FC } from "react";
import MonacoEditor from "@monaco-editor/react";
// import { useTheme } from "../../../../../util/store/theme";

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

const Editor: FC<Omit<EditorProps, "beforeMount" | "onMount">> = ({
  options,
  ...props
}) => (
  /* const theme = useTheme();*/ <MonacoEditor
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
    // theme={theme === "dark" ? "direktiv-dark" : "vs-light"}
    theme="direktiv-dark"
    {...props}
  />
);

export default Editor;
