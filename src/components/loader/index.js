import { useEffect, useState } from "react"
import './style.css'
let timeout = null
export default function Loader(props) {

    const {load, children, timer} = props
    
    const [display, setDisplay] = useState(false)
    const [ show, setShow ] = useState(true)
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
            clearTimeout(timeout)
            timeout = setTimeout(()=>{
                setDisplay(true)
            },timer)
        }
    },[children])
    
    useEffect(()=>{
        if(load === false)
            setTimeout(()=>{
                setShow(false)
            }, 1000)
        if(load === true)
            setTimeout(()=>{
                setShow(true)
            }, 0) 
    }, [load])
    if(show) {
        // return a loader
        return (
            <div className="container" style={{display: display ? "none": "flex"}}>
                <div className="loader" >
                </div>
            </div>
        )
    }

    return(
        children
    )
}