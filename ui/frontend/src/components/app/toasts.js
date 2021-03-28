import React, {useEffect, useState} from 'react'
import 'css/namespace.css'


import Toast from 'react-bootstrap/Toast'
import ProgressBar from 'react-bootstrap/ProgressBar'


import {TimeSince as calcTimeSince} from 'util/utils'

export default function CustomToast(props) {
    const {destroyCallback, title, body, timeStamp} = props
    const [showA, setShowA] = useState(true);
    const [showProgress, setShowProgress] = useState(true);
    const [progress, setProgress] = useState(0);
    const [timeSince, setTimeSince] = useState("now");

    useEffect(() => {
        var myProgress = progress
        let interval = setInterval(() => {
            if (myProgress >= 105 && showProgress) {
                setShowA(false);
                if (destroyCallback) {
                    destroyCallback(timeStamp)
                }
                clearInterval(interval)
            }

            setProgress(progress => progress + 2);

            // Hacky way to access showProgress inside interval
            setShowProgress(showProgress => {
                if (!showProgress) {
                    clearInterval(interval)
                }
                return showProgress
            })
            myProgress = myProgress + 2
        }, 100);

        return () => {
            clearInterval(interval);
        };
    }, [destroyCallback, progress, showProgress, timeStamp]);

    useEffect(() => {
        let isSubscribed = true
        let timeStampRate = 1000

        setInterval(() => {
            if (isSubscribed) {
                setTimeSince(calcTimeSince(timeStamp))
            }
        }, timeStampRate)
        return () => isSubscribed = false
    }, [timeStamp]);

    return (
        <Toast show={showA} onClose={() => {
            setShowA(false)
            destroyCallback(timeStamp)
        }} onMouseMove={() => {
            if (showProgress) {
                setShowProgress(false)
            }
        }}>
            <Toast.Header style={{
                width: '24em',
            }}>
                {/* <img
                    src="holder.js/20x20?text=%20"
                    className="rounded mr-2"
                    alt=""
                /> */}
                <strong className="mr-auto">{title}</strong>
                <small>{timeSince}</small>
            </Toast.Header>
            <Toast.Body style={{padding: "0"}}>
                {showProgress ? (<ProgressBar style={{borderRadius: "0px", height: "2px"}} now={progress}/>) : (<></>)}
                <div style={{padding: "0.75em"}}>
                    {body}
                </div>
            </Toast.Body>
        </Toast>
    )
}