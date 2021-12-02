import Editor, {useMonaco} from "@monaco-editor/react";
import { useEffect } from "react";
import './style.css'

export default function DirektivEditor(props) {
    const {dvalue, dlang, value, height, width, setDValue, onMount, readonly} = props
    
    const monaco = useMonaco()

    useEffect(()=>{
        // console.log(monaco)
        // if(monaco !== null) {
        //     monaco.editor.colorize(dvalue, dlang)
        // }
        // monaco.editor.layout()
    },[monaco, dlang])

    function handleEditorChange(value, event) {
        setDValue(value)
    }

    return (
        <Editor
            options={{
                readOnly: readonly
            }}
            height={height}
            width={width}
            language={dlang}
            defaultValue={dvalue}
            value={value}
            theme={"vs-dark"}
            loading={"This shows when component is loading"}
            onChange={handleEditorChange}
            onMount={onMount}
        />
    )
}