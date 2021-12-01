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

    // when children change reset the timer
    useEffect(()=>{
        if(display){
            setDisplay(false)
            setTimeout(()=>{
                setDisplay(true)
            },timer)
        }
    },[children])

    if(load) {
        // return a loader
        
        return (
            <div style={{ display:"flex", alignItems:"center", justifyContent:"center", flex: 1, width:"100%", height:"100%", padding:"8px"}}>
                <div style={{visibility: display ? "visible": "hidden"}} className="loader">
                </div>
            </div>

        )
    }

    return(
        children
    )
}