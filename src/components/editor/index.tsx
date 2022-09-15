import Editor, {useMonaco, loader, EditorProps } from "@monaco-editor/react";
import { useEffect, useState } from "react";
import './style.css'
// import * as cobalt from './cobalt.json'

loader.config({ paths: { vs: "/monaco-0.28.1-min/vs" } });

const cobalt = {
    "base": "vs-dark",
    "inherit": true,
    "rules": [
      {
        "background": "002240",
        "token": ""
      },
      {
        "foreground": "e1efff",
        "token": "punctuation - (punctuation.definition.string || punctuation.definition.comment)"
      },
      {
        "foreground": "ff628c",
        "token": "constant"
      },
      {
        "foreground": "ffdd00",
        "token": "entity"
      },
      {
        "foreground": "ff9d00",
        "token": "keyword"
      },
      {
        "foreground": "ffee80",
        "token": "storage"
      },
      {
        "foreground": "3ad900",
        "token": "string -string.unquoted.old-plist -string.unquoted.heredoc"
      },
      {
        "foreground": "3ad900",
        "token": "string.unquoted.heredoc string"
      },
      {
        "foreground": "0088ff",
        "fontStyle": "italic",
        "token": "comment"
      },
      {
        "foreground": "80ffbb",
        "token": "support"
      },
      {
        "foreground": "cccccc",
        "token": "variable"
      },
      {
        "foreground": "ff80e1",
        "token": "variable.language"
      },
      {
        "foreground": "ffee80",
        "token": "meta.function-call"
      },
      {
        "foreground": "f8f8f8",
        "background": "800f00",
        "token": "invalid"
      },
      {
        "foreground": "ffffff",
        "background": "223545",
        "token": "text source"
      },
      {
        "foreground": "ffffff",
        "background": "223545",
        "token": "string.unquoted.heredoc"
      },
      {
        "foreground": "ffffff",
        "background": "223545",
        "token": "source source"
      },
      {
        "foreground": "80fcff",
        "fontStyle": "italic",
        "token": "entity.other.inherited-class"
      },
      {
        "foreground": "9eff80",
        "token": "string.quoted source"
      },
      {
        "foreground": "80ff82",
        "token": "string constant"
      },
      {
        "foreground": "80ffc2",
        "token": "string.regexp"
      },
      {
        "foreground": "edef7d",
        "token": "string variable"
      },
      {
        "foreground": "ffb054",
        "token": "support.function"
      },
      {
        "foreground": "eb939a",
        "token": "support.constant"
      },
      {
        "foreground": "ff1e00",
        "token": "support.type.exception"
      },
      {
        "foreground": "8996a8",
        "token": "meta.preprocessor.c"
      },
      {
        "foreground": "afc4db",
        "token": "meta.preprocessor.c keyword"
      },
      {
        "foreground": "73817d",
        "token": "meta.sgml.html meta.doctype"
      },
      {
        "foreground": "73817d",
        "token": "meta.sgml.html meta.doctype entity"
      },
      {
        "foreground": "73817d",
        "token": "meta.sgml.html meta.doctype string"
      },
      {
        "foreground": "73817d",
        "token": "meta.xml-processing"
      },
      {
        "foreground": "73817d",
        "token": "meta.xml-processing entity"
      },
      {
        "foreground": "73817d",
        "token": "meta.xml-processing string"
      },
      {
        "foreground": "9effff",
        "token": "meta.tag"
      },
      {
        "foreground": "9effff",
        "token": "meta.tag entity"
      },
      {
        "foreground": "9effff",
        "token": "meta.selector.css entity.name.tag"
      },
      {
        "foreground": "ffb454",
        "token": "meta.selector.css entity.other.attribute-name.id"
      },
      {
        "foreground": "5fe461",
        "token": "meta.selector.css entity.other.attribute-name.class"
      },
      {
        "foreground": "9df39f",
        "token": "support.type.property-name.css"
      },
      {
        "foreground": "f6f080",
        "token": "meta.property-group support.constant.property-value.css"
      },
      {
        "foreground": "f6f080",
        "token": "meta.property-value support.constant.property-value.css"
      },
      {
        "foreground": "f6aa11",
        "token": "meta.preprocessor.at-rule keyword.control.at-rule"
      },
      {
        "foreground": "edf080",
        "token": "meta.property-value support.constant.named-color.css"
      },
      {
        "foreground": "edf080",
        "token": "meta.property-value constant"
      },
      {
        "foreground": "eb939a",
        "token": "meta.constructor.argument.css"
      },
      {
        "foreground": "f8f8f8",
        "background": "000e1a",
        "token": "meta.diff"
      },
      {
        "foreground": "f8f8f8",
        "background": "000e1a",
        "token": "meta.diff.header"
      },
      {
        "foreground": "f8f8f8",
        "background": "4c0900",
        "token": "markup.deleted"
      },
      {
        "foreground": "f8f8f8",
        "background": "806f00",
        "token": "markup.changed"
      },
      {
        "foreground": "f8f8f8",
        "background": "154f00",
        "token": "markup.inserted"
      },
      {
        "background": "8fddf630",
        "token": "markup.raw"
      },
      {
        "background": "004480",
        "token": "markup.quote"
      },
      {
        "background": "130d26",
        "token": "markup.list"
      },
      {
        "foreground": "c1afff",
        "fontStyle": "bold",
        "token": "markup.bold"
      },
      {
        "foreground": "b8ffd9",
        "fontStyle": "italic",
        "token": "markup.italic"
      },
      {
        "foreground": "c8e4fd",
        "background": "001221",
        "fontStyle": "bold",
        "token": "markup.heading"
      }
    ],
    "colors": {
      "editor.foreground": "#FFFFFF",
      "editor.background": "#002240",
      "editor.selectionBackground": "#B36539BF",
      "editor.lineHighlightBackground": "#00000059",
      "editorCursor.foreground": "#FFFFFF",
      "editorWhitespace.foreground": "#FFFFFF26"
    }
}

export interface DirektivEditorProps extends React.HTMLAttributes<HTMLDivElement> {
  /**
  * Callback function for CRTL+S shortcut is pressed when editor is in focus. 
  */
  saveFn?: () => void
  /**
  * Removes border radius of editor component.
  */
  noBorderRadius?: boolean
  /**
  * Additional Monaco Editor options
  */
  options?: EditorProps
  /**
  * Default value for editor to display.
  */
  dvalue: EditorProps["defaultValue"]
  /**
  * Monaco supported Language for editor to use. Example: 'json', 'yaml'
  */
  dlang: string
  /**
  * Value of editor's content.
  */
  value: EditorProps["value"]
  /**
  * Editor Height. Editor height is calculated by subtracting 18 to this value.
  */
  height?: number
  /**
  * Editor Width.
  */
  width?: number
  setDValue?: React.Dispatch<React.SetStateAction<string>>
  /**
  * Sets editor to Read Only mode.
  */
  readonly?: boolean
  /**
  * Toggle editor validation on value.
  */
  validate?: boolean
  /**
  * Toggles minimap in editor.
  */
  minimap?: boolean
}

// DirektivEditor - Eidtor Component
// Note: width and height must not have unit suffix. e.g. 400=acceptable, 400% will not work
// TODO: Support multiple width/height unit

/**
* Logs component that only renders the visible log items of a list. 
* Supports auto scrolling to bottom as logs change, and word wrapping.
*/
function DirektivEditor({
    saveFn,
    noBorderRadius,
    options,
    dvalue,
    dlang = "json",
    value,
    height,
    width,
    setDValue,
    readonly,
    validate,
    minimap,
 ...props 
}: DirektivEditorProps) {
    const monaco = useMonaco()

    const [ed, setEditor] = useState(null as any);

    useEffect(()=>{
        if(monaco) {           
            monaco.editor.defineTheme('cobalt', cobalt)
            monaco.editor.setTheme('cobalt')
            if (monaco?.languages?.[dlang]?.[`${dlang}Defaults`]?.setDiagnosticsOptions) {
              monaco.languages[dlang][`${dlang}Defaults`].setDiagnosticsOptions({
                validate: validate === undefined ? true : validate,
              })
            } else {
              console.warn(`editor warning: ${dlang} does not support Diagnostics`)
            }

        }
    },[monaco, dlang, validate])

    function handleEditorChange (value: string | undefined) {
        if (value && setDValue) {
            setDValue(value)
        }
    }

    function setCommonEditorTriggers(editor: any, monaco: any) {
      editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KEY_J, ()=>{
        editor.trigger('fold', 'editor.foldAll')
      })
      editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KEY_K, ()=>{
        editor.trigger('unfold', 'editor.unfoldAll')
      })
    }


    if (saveFn && ed) {
      ed.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KEY_S, saveFn)
    }

    let handleEditorDidMount = function(editor: any, monaco: any) {
      setCommonEditorTriggers(editor, monaco)
      setEditor(editor)
    }

    if (readonly) {
      handleEditorDidMount = function(editor, monaco) {
        setCommonEditorTriggers(editor, monaco)
        let messageContribution = editor.getContribution('editor.contrib.messageController');
        editor.onDidAttemptReadOnlyEdit(() => {
          messageContribution.dispose();
        })
      }
    }

    return (
      <div {...props} className={`monaco-editor monaco-wrapper ${props.className ? props.className : ""}`} style={{ borderRadius: !noBorderRadius ? "8px" : "0px", width: width, height: height ? height-18 : height, ...props.style}}>
        <Editor
            options={{
                ...options,
                readOnly: readonly,
                scrollBeyondLastLine: false,
                cursorBlinking: "smooth",
                quickSuggestions: false,
                minimap: {
                  enabled: minimap === undefined ? false : minimap,
                },
                wordWrap: "wordWrapColumn",
                wordWrapColumn: 10000
              }}
            height={height ? height-18 : height}
            width={width}
            language={dlang}
            defaultValue={dvalue}
            value={value}
            theme={"cobalt"}
            loading={"Loading component..."}
            onChange={handleEditorChange}
            onMount={handleEditorDidMount}
        />
      </div>
    )
}

export default DirektivEditor