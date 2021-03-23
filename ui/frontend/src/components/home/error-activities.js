import React from 'react'
import 'css/activities.css'

export default function ErrorActivitiesArea(props) {
    const {error} = props
    return (
        <div id="empty-activities-area">
            <p style={{color: 'red', marginBottom: "0px"}}>
                Error fetching instances: "{error}"
            </p>
        </div>
    );
}