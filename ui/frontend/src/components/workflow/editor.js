import React, {useEffect, useState} from "react";
import "css/workflow.css";

// Code Mirror
import {Controlled as CodeMirror} from "react-codemirror2";
import "codemirror/lib/codemirror.css";

// custard theme
import "css/editor.css";

// themes
import "codemirror/theme/idea.css";
import "codemirror/addon/lint/lint.js";
import "codemirror/addon/lint/lint.css";
import "codemirror/addon/hint/show-hint.css";

// YAML
import "codemirror/mode/yaml/yaml.js";
import "codemirror/addon/lint/yaml-lint";
import YAML from "js-yaml";

// Javascript (JSON)
import "codemirror/mode/javascript/javascript.js";
import "codemirror/addon/lint/javascript-lint";
import "codemirror/addon/hint/javascript-hint";
import {JSHINT} from "jshint";

// EXTRA
window.JSHINT = JSHINT;
window.jsyaml = YAML;

const EDITOR_HEIGHT = 500

export default function Editor(props) {
    const [editorInfo, setEditorInfo] = useState({init: false, editor: null});
    const {
        style,
        value,
        onChange,
        mode,
        readOnly
    } = props;

    let rOnly = true
    if (readOnly != undefined) {
        rOnly = readOnly
    }

    useEffect(() => {
        if (!editorInfo.init && editorInfo.editor) {
            editorInfo.editor.setSize(null, `${EDITOR_HEIGHT - 5}`);
            setEditorInfo((editorInfo) => {
                editorInfo.init = true;
                return editorInfo;
            });
        }
    }, [editorInfo]);

    return (
        <div style={{border: "solid 2px #777777", ...style}}>
            <div style={{height: `${EDITOR_HEIGHT}px`}}>
                <CodeMirror
                    editorDidMount={(editor) => {
                        setEditorInfo({init: false, editor: editor});
                        if (mode === "json") {
                            onChange("{\n  \n}");
                        }
                    }}
                    className="editor"
                    value={value}
                    onBeforeChange={(editor, data, value) => {
                        onChange(value);
                    }}
                    onBlur={(editor, data, value) => {
                    }}
                    options={{
                        cursorBlinkRate: rOnly ? -1 : 530,
                        autoRefresh: true,
                        gutters: ["CodeMirror-lint-markers"],
                        mode: mode === "json" ? "javascript" : "yaml",
                        theme: "idea",
                        lineNumbers: true,
                        lineWrapping: true,
                        lint: true,
                        readOnly: rOnly,
                    }}
                />
            </div>
        </div>
    );
}
