import { useEffect, useState } from "react"
import './style.css'
export default function Loader(props) {

    const {load, children, timer} = props
    
    const [display, setDisplay] = useState(false)

    useEffect(()=>{
        if(timer){
            setTimeout(()=>{
                setDisplay(true)
            },timer)
        }
    },[timer])


    // display blank like nothing has changed while its currently loading
    if(!display){
        return ""
    }

    if(load) {
        // return a loader
        
        return (
            <div style={{display:"flex", alignItems:"center", justifyContent:"center", flex: 1}}>
            <div className="loader">
            </div>
            </div>

        )
    }

    return(
        children
    )
}