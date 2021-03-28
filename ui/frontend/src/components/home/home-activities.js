import React, {useCallback, useContext, useEffect, useState} from 'react'
import {Link} from 'react-router-dom'
import logoColor from "img/logo-color.png";

import ServerContext from 'components/app/context'
import EmptyActivitiesArea from 'components/home/empty-activities'
import ErrorActivitesArea from './error-activities'

import {InstanceBadge, InstanceStatusParse, TimeSinceUnix} from 'util/utils'

export default function HomepageActivities(props) {

    let namespaces = props.namespaces
    const context = useContext(ServerContext);
    const [loader, setLoader] = useState(false)
    const [instances, setInstances] = useState([]);
    const [err, setError] = useState("")

    const fetchInstances = useCallback(
        () => {
            async function fetchInstances(name) {
                try {
                    let resp = await context.Fetch(`/instances/${name}`, {
                        method: `GET`
                    })
                    if (resp.ok) {
                        let json = await resp.json()
                        let wfI = json.workflowInstances
                        let out = [];
                        for (let i = 0; i < wfI.length; i++) {
                            out.push({
                                namespace: name,
                                id: wfI[i].id,
                                created: wfI[i].beginTime.seconds,
                                status: wfI[i].status,
                            })
                        }
                        return out
                    } else {
                        throw (new Error(await resp.text()))
                    }

                } catch (e) {
                    setError(e.message)
                }

            }

            async function fetchAllInstances() {
                setLoader(true)
                let arr = []
                for (let i = 0; i < namespaces.length; i++) {
                    // checking if it returns undefined
                    let x = await fetchInstances(namespaces[i].name)
                    if (x) {
                        arr = arr.concat(x)
                    }
                }
                setInstances(arr)
                setLoader(false)
            }

            fetchAllInstances()
        },
        [namespaces, context.Fetch],
    )

    useEffect(() => {
        fetchInstances()
    }, [namespaces, fetchInstances])

    if (err !== "") {
        return <ErrorActivitesArea error={err}/>
    } else if (instances.length === 0) {
        return <EmptyActivitiesArea/>
    } else {
        return (
            <>
                {loader ?
                    <div id="instances">

                        <div style={{
                            minHeight: "500px",
                            display: "flex",
                            alignItems: "center",
                            justifyContent: "center"
                        }}>
                            <img
                                alt="loading symbol"
                                src={logoColor}
                                height={200}
                                className="animate__animated animate__bounce animate__infinite"/>
                        </div>

                    </div>
                    :
                    <div id="instances">
                        <div style={{display: 'flex'}}>
                            <h5 style={{
                                flex: 1,
                                justifyContent: "center",
                                alignContent: "center",
                                alignSelf: "center"
                            }}>
                                Instances
                            </h5>
                            <div style={{flex: 1, textAlign: "right", justifyContent: "center", marginBottom: ".5rem"}}>
                            </div>
                        </div>
                        {instances.map((obj) => <InstanceListItem key={obj.id} instanceData={obj}/>)}
                    </div>
                }
            </>
        );
    }


}

function InstanceListItem(props) {
    let {namespace, id, status, created} = props.instanceData;
    let timeSince = TimeSinceUnix(created)

    let classes = "instance-list-item";
    let pStatus = InstanceStatusParse(status)
    if (pStatus) {
        classes = `${classes} ${pStatus}`
    }

    // Gather the workflow name from the ID
    let workflow = id.split("/")[1]

    // Gather the instance name from the ID
    let instance = id.split("/")[2]


    return (
        <>
            <Link to={`/i/${id}`}>
                <div className={classes} style={{display: 'flex'}}>
                    <div className="instance-list-item-title" style={{flex: 1}}>
                        <div>
                            <b>Workflow:</b> {workflow}<br/>
                            <b>Instance:</b> {instance}<br/>
                            <b>Namespace:</b> {namespace}

                        </div>
                        <div style={{textAlign: "right"}}>
                        </div>
                    </div>
                    <div style={{textAlign: "right"}}>
                        <div>
                            {`${timeSince}`}
                        </div>
                        <div>
                            {InstanceBadge(pStatus)}
                        </div>
                    </div>
                </div>
            </Link>

        </>

    );
}