import Editor, {useMonaco} from "@monaco-editor/react";
import { useEffect } from "react";

export default function DirektivEditor(props) {
    const {dvalue, dlang, value, height, width, setDValue, onMount} = props
    
    const monaco = useMonaco()

    useEffect(()=>{
        console.log(monaco)
        // monaco.editor.layout()
    },[monaco])

    function handleEditorChange(value, event) {
        setDValue(value)
    }

    return (
        <Editor
            height={height}
            width={width}
            defaultLanguage={dlang}
            defaultValue={dvalue}
            value={value}
            theme={"vs-dark"}
            loading={"This shows when component is loading"}
            onChange={handleEditorChange}
            onMount={onMount}
        />
    )
}