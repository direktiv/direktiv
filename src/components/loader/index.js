import { useEffect, useState } from "react"
import './style.css'


export default function Loader(props) {

    const {load, children, timer} = props
    const [display, setDisplay] = useState(false)
    const [timeoutTimer, setTimeoutTimer] = useState(null)

    console.log(load, children, timer, display)

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


    // useEffect(()=>{
    //     if(timer){
    //         let t = setTimeout(()=>{
    //             setDisplay(true)
    //         },timer)
    //     }
    // },[timer])

    // useEffect(()=>{
    //     return () => {
    //         if(TimeoutTimer) {
    //             clearTimeout
    //         }
    //     }
    // },[])

    // useEffect(()=>{
    //     if(display){
    //         setDisplay(false)
    //         setTimeout(()=>{
    //             setDisplay(true)
    //         },timer)
    //     }
    // },[children, display, timer])


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