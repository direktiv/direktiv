import Editor from "@monaco-editor/react";

export default function DirektivEditor(props) {
    const {dvalue, dlang, height, width, setDValue, onMount} = props
    
    function handleEditorChange(value, event) {
        setDValue(value)
    }

    return (
        <Editor
            height={height}
            width={width}
            defaultLanguage={dlang}
            defaultValue={dvalue}
            theme={"vs-dark"}
            loading={"This shows when component is loading"}
            onChange={handleEditorChange}
            onMount={onMount}
        />
    )
}