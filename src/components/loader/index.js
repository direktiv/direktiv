import { useEffect, useState } from "react"
import './style.css'


export default function Loader(props) {

    const {load, children, timer} = props
    const [display, setDisplay] = useState(false)
    const [timeoutTimer, setTimeoutTimer] = useState(null)

    // show loader if timer is hit set timeout to set display to true
    useEffect(()=>{
        if(timer !== null && load) {
            let t = setTimeout(()=>{
                setDisplay(true)
            }, timer)
            setTimeoutTimer(t)
        }
    },[timer, load])

    // check if load has been changed to true
    useEffect(()=>{
        // if its finished loading and timeoutTimer isn't null show children
        if(!load && timeoutTimer !== null){
            clearTimeout(timeoutTimer)
            setDisplay(false)
        }
    },[load, timeoutTimer])

    if(display && load) {
        // return a loader
        return (
            <div className="container" style={{display:'flex'}}>
                <div className="loader" >
                </div>
            </div>
        )
    }
    
    if(load){
        return ""
    }

    return(
        children
    )
}