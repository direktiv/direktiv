import { useEffect, useState } from "react"
import './style.css'

export interface LoaderProps {
    /**
    * Children Loader will overlay
    */
    children?: React.ReactNode
    /**
    * Milliseconds to wait before, displaying animated loader
    */
    timer: number
    /**
    * Whether or not component is currently loading
    */
    load: boolean
}

/**
*  Wrapper over React Node children that will display an animated loading icon if load=true. Otherwise the children are returned. 
*/
function Loader({ children, timer=1000, load }: LoaderProps) {

    const [display, setDisplay] = useState(false)
    const [timeoutTimer, setTimeoutTimer] = useState(undefined as NodeJS.Timeout | undefined)

    // show loader if timer is hit set timeout to set display to true
    useEffect(() => {
        if (load) {
            let t = setTimeout(() => {
                setDisplay(true)
            }, timer)
            setTimeoutTimer(t)
        }
    }, [timer, load])

    // check if load has been changed to true
    useEffect(() => {
        // if its finished loading and timeoutTimer isn't null show children
        if (!load && !timeoutTimer) {
            clearTimeout(timeoutTimer)
            setDisplay(false)
        }
    }, [load, timeoutTimer])

    if (display && load) {
        // return a loader
        return (
            <div className="container" style={{ display: 'flex' }}>
                <div className="loader" >
                </div>
            </div>
        )
    }

    if (load) {
        return <></>
    }

    return (
        <>
            {children}
        </>
    )
}

export default Loader