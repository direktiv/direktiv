import Editor from "@monaco-editor/react";

export default function DirektivEditor(props) {
    const {dvalue, dlang, height, width, setDValue} = props
    
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
        />
    )
}