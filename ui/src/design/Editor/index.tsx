import * as monaco from "monaco-editor";

import { FC, useRef } from "react";
import MonacoEditor, { loader } from "@monaco-editor/react";

import AutoSizer from "react-virtualized-auto-sizer";
import type { EditorProps } from "@monaco-editor/react";
import { supportedLanguages } from "./utils";
import themeDark from "./theme-dark";
import themeLight from "./theme-light";

loader.config({ monaco });

const beforeMount: EditorProps["beforeMount"] = (monaco) => {
  monaco.editor.defineTheme("direktiv-dark", themeDark);
  monaco.editor.defineTheme("direktiv-light", themeLight);
};

export type EditorLanguagesType = (typeof supportedLanguages)[number];

type EditorType = Parameters<NonNullable<EditorProps["onMount"]>>[0];

const Editor: FC<
  Omit<EditorProps, "beforeMount" | "onMount" | "onChange"> & {
    theme?: "light" | "dark";
    onSave?: (value: string | undefined) => void;
    onChange?: (value: string | undefined) => void;
    onMount?: EditorProps["onMount"];
    language?: EditorLanguagesType;
  }
> = ({
  options,
  theme,
  onSave,
  onChange,
  onMount,
  language = "yaml",
  ...props
}) => {
  const monacoRef = useRef<EditorType>();

  const handleChange = () => {
    onChange && onChange(monacoRef.current?.getValue());
  };

  // this is the shared onMount function, that will be called for
  // every Editor component. Each Editor can implement their own
  // onMount function on top of this one.
  const commonOnMount: EditorProps["onMount"] = (editor, monaco) => {
    monacoRef.current = editor;
    onMount?.(editor, monaco);
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
          onMount={commonOnMount}
          onChange={() => {
            handleChange();
          }}
          options={{
            // options reference: https://microsoft.github.io/monaco-editor/typedoc/interfaces/editor.IEditorOptions.html
            scrollBeyondLastLine: false,
            cursorBlinking: "smooth",
            wordWrap: "on",
            fontSize: 13,
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
