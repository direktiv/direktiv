import type { EditorProps } from "@monaco-editor/react";
import { FC } from "react";
import MonacoEditor from "@monaco-editor/react";
import { useTheme } from "../../../../../util/store/theme";

const Editor: FC<EditorProps> = ({ ...props }) => {
  const theme = useTheme();
  return (
    <MonacoEditor
      options={{
        scrollBeyondLastLine: false,
        cursorBlinking: "smooth",
        wordWrap: true,
        minimap: {
          enabled: false,
        },
      }}
      loading=""
      language="yaml"
      theme={theme === "dark" ? "vs-dark" : "vs-light"}
      {...props}
    />
  );
};

export default Editor;
