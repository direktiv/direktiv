import { FC, useRef } from "react";

import AutoSizer from "react-virtualized-auto-sizer";
import type { EditorProps } from "@monaco-editor/react";
import MonacoEditor from "@monaco-editor/react";
import themeDark from "./theme-dark";
import themeLight from "./theme-light";

const beforeMount: EditorProps["beforeMount"] = (monaco) => {
  monaco.editor.defineTheme("direktiv-dark", themeDark);
  monaco.editor.defineTheme("direktiv-light", themeLight);
};

type EditorType = Parameters<NonNullable<EditorProps["onMount"]>>[0];

export type EditorLanguagesType =
  | "html"
  | "css"
  | "json"
  | "shell"
  | "plaintext"
  | "yaml";

const Editor: FC<
  Omit<EditorProps, "beforeMount" | "onMount" | "onChange"> & {
    theme?: "light" | "dark";
    onSave?: (value: string | undefined) => void;
    onChange?: (value: string | undefined) => void;
    language?: EditorLanguagesType;
  }
> = ({ options, theme, onSave, onChange, language = "yaml", ...props }) => {
  const monacoRef = useRef<EditorType>();

  const handleChange = () => {
    onChange && onChange(monacoRef.current?.getValue());
  };

  const onMount: EditorProps["onMount"] = (editor, monaco) => {
    monacoRef.current = editor;
    editor.focus();
    monacoRef.current.onDidChangeModelContent(handleChange);

    onSave &&
      editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
        onSave(
          monacoRef.current?.getValue()
            ? `${monacoRef.current?.getValue()}`
            : undefined
        );
      });
  };

  return (
    <AutoSizer>
      {({ height, width }) => (
        <MonacoEditor
          // remove "Cannot edit in read-only editor" tooltip
          className="[&_.monaco-editor-overlaymessage]:!hidden"
          width={width}
          height={height}
          beforeMount={beforeMount}
          onMount={onMount}
          options={{
            // options reference: https://microsoft.github.io/monaco-editor/typedoc/interfaces/editor.IEditorOptions.html
            scrollBeyondLastLine: false,
            cursorBlinking: "smooth",
            wordWrap: true,
            fontSize: "13px",
            minimap: {
              enabled: false,
            },
            contextmenu: false,
            ...options,
          }}
          loading=""
          language={language}
          theme={theme === "dark" ? "direktiv-dark" : "direktiv-light"}
          {...props}
        />
      )}
    </AutoSizer>
  );
};

export default Editor;
