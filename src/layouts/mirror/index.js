import Tippy from '@tippyjs/react';
import { useMirror, useNodes } from 'direktiv-react-hooks';
import React, { useEffect, useRef, useState } from 'react';
import { VscAdd, VscLock, VscSync, VscUnlock } from 'react-icons/vsc';
import { useNavigate, useParams } from 'react-router';
import Alert from '../../components/alert';
import Button from '../../components/button';
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import FlexBox from '../../components/flexbox';
import Loader from '../../components/loader';
import { ButtonDefinition, ModalHeadless } from '../../components/modal';
import { Config } from '../../util';
import ActivityTable from './activities';
import MirrorInfoPanel from './info';
import ActivityLogs from './logs';
import './style.css';


export default function MirrorPage(props) {
    const { namespace, setBreadcrumbChildren } = props
    const params = useParams()
    const navigate = useNavigate()
    const [activity, setActivity] = useState(null)
    const [currentlyLocking, setCurrentlyLocking] = useState(true)
    const [isReadOnly, setIsReadOnly] = useState(true)

    const [errorMsg, setErrorMsg] = useState(null)
    const [load, setLoad] = useState(true)
    const [syncVisible, setSyncVisible] = useState(false)

    let path = `/`
    if (params["*"] !== undefined) {
        path = `/${params["*"]}`
    }

    const { info, activities, err, setLock, updateSettings, cancelActivity, sync } = useMirror(Config.url, true, namespace, path, localStorage.getItem("apikey"), "limit=50", "order.field=CREATED", "order.direction=DESC")
    const { data, getNode, err: nodeErr } = useNodes(Config.url, false, namespace, path, localStorage.getItem("apikey"), `limit=1`)

    const setLockRef = useRef(setLock)
    const syncRef = useRef(sync)
    const getNodeRef = useRef(getNode)
    const setBreadcrumbChildrenRef = useRef(setBreadcrumbChildren)


    // Error Handling Non existent node and bad mirror
    useEffect(() => {
        if (err) {
            setErrorMsg("Error getting mirror info: " + err)
        } else if (nodeErr) {
            console.error("Error getting node: ", nodeErr)
            navigate(`/n/${namespace}/explorer${path}`)
        }
    }, [nodeErr, err, data, navigate, namespace, path])

    // Error Handling bad node
    useEffect(() => {
        if (!getNodeRef.current) {
            return
        }

        if (!load && data) {
            getNodeRef.current().then((nodeData) => {
                if (nodeData.node.expandedType !== "git") {
                    navigate(`/n/${namespace}/explorer${path}`)
                }
            }).catch((e) => {
                navigate(`/n/${namespace}/explorer${path}`)
            })
        }
    }, [data, load, navigate, namespace, path])

    // Keep track of getNodeRef
    useEffect(() => {
        getNodeRef.current = getNode
    }, [getNode])


    useEffect(() => {
        if (nodeErr) {
            setErrorMsg("Error getting node: " + nodeErr)
            return
        }

        const handler = setTimeout(() => {
            if (currentlyLocking) {
                getNode().then((nodeData) => {
                    setIsReadOnly(nodeData.node.readOnly)
                }).catch((e) => {
                    setErrorMsg("Error getting node: " + e.message)
                }).finally(() => {
                    setCurrentlyLocking(false)
                })
            }
        }, 1000)

        return () => {
            clearTimeout(handler);
        };
    }, [currentlyLocking, getNode, nodeErr])

    useEffect(() => {
        if (data && info) {
            setLoad(false)
        }
    }, [data, info, load])

    useEffect(() => {
        if (!setBreadcrumbChildrenRef.current || !syncRef.current) {
            return
        }

        setBreadcrumbChildrenRef.current((
            <FlexBox className="center row gap" style={{ justifyContent: "flex-end", paddingRight: "6px" }}>
                <Button id="btn-sync-mirror" tip={"Sync mirror with remote"} disabledTip={"Cannot sync mirror while Writable"} disabled={!isReadOnly} className="small light bold shadow" style={{ fontWeight: "bold", width: "fit-content" }} onClick={()=>{
                    setSyncVisible(!syncVisible)
                }}>
                    <FlexBox className="row center gap-sm">
                        <VscSync />
                        Sync
                    </FlexBox>
                </Button>
                <ModalHeadless
                    visible={syncVisible}
                    setVisible={setSyncVisible}
                    escapeToCancel
                    activeOverlay
                    title="Sync Mirror"
                    titleIcon={
                        <VscSync />
                    }
                    style={{
                        maxWidth: "68px"
                    }}
                    modalStyle={{
                        width: "300px"
                    }}
                    actionButtons={[
                        ButtonDefinition("Sync", async () => {
                            await syncRef.current(true)
                        }, "small", () => { }, true, false),
                        ButtonDefinition("Cancel", () => { }, "small light", () => { }, true, false)
                    ]}
                >
                    <FlexBox className="col gap" style={{ paddingTop: "8px" }}>
                        <FlexBox className="col center info-update-label">
                          Fetch and sync mirror with latest content from remote repository?
                        </FlexBox>
                    </FlexBox>
                </ModalHeadless>
                <Button className={`small light bold shadow ${currentlyLocking ? "loading disabled" : ""}`} style={{ fontWeight: "bold", width: "fit-content", whiteSpace: "nowrap" }} onClick={async () => {
                    if (isReadOnly) {
                        setCurrentlyLocking(true)

                        try {
                            await setLockRef.current(true)
                        } catch (e) {
                            setCurrentlyLocking(false)
                            setErrorMsg(e.message)
                        }
                    } else {
                        setCurrentlyLocking(true)
                        try {
                            await setLockRef.current(false)
                        } catch (e) {
                            setCurrentlyLocking(false)
                            setErrorMsg(e.message)
                        }
                    }
                }}>
                    <FlexBox className="row center gap-sm">
                        {isReadOnly ?
                            <>

                                <VscUnlock />
                                Make Writable
                            </>
                            :
                            <>
                                <VscLock />
                                Make ReadOnly
                            </>
                        }
                    </FlexBox>
                </Button>
                {isReadOnly ? <MirrorReadOnlyBadge /> : <MirrorWritableBadge />}
            </ FlexBox>
        ))
    }, [currentlyLocking, isReadOnly, syncVisible])

    // Keep Refs up to date
    useEffect(() => {
        setBreadcrumbChildrenRef.current = setBreadcrumbChildren
        setLockRef.current = setLock
        syncRef.current = sync
    }, [setBreadcrumbChildren, setLock, sync])


    // Unmount cleanup breadcrumb children
    useEffect(() => {
        return (() => {
            if (setBreadcrumbChildrenRef.current) {
                setBreadcrumbChildrenRef.current(<></>)
            }
        })
    }, [])

    if (!namespace) {
        return <></>
    }


    return (
        <>
            <Loader load={load} timer={1000}>
                {
                    errorMsg ?
                        <FlexBox style={{ maxHeight: "50px", paddingRight: "6px", paddingBottom: "8px" }}>
                            <Alert setErrorMsg={setErrorMsg} className="critical" style={{ height: "100%" }}>{`Error: ${errorMsg}`}</Alert>
                        </FlexBox>
                        : <></>
                }
                <FlexBox className="col gap" style={{ paddingRight: "8px" }}>
                    {/* <BreadcrumbCorner>
                    </BreadcrumbCorner> */}
                    <FlexBox className="row gap wrap" style={{ flex: "1 1 0%", maxHeight: "65vh" }}>
                        <ContentPanel id={`panel-activity-list`} style={{ flex: 2, width: "100%", minHeight: "60vh", maxHeight: "55vh" }}>
                            <ContentPanelTitle>
                                <ContentPanelTitleIcon>
                                    <VscAdd />
                                </ContentPanelTitleIcon>
                                <FlexBox className="gap" style={{ alignItems: "center" }}>Activity List</FlexBox>
                            </ContentPanelTitle>
                            <ContentPanelBody style={{ overflow: "auto" }}>
                                <FlexBox style={{ flexShrink: "1", height: "fit-content" }}>
                                    <ActivityTable activities={activities} setActivity={setActivity} cancelActivity={cancelActivity} setErrorMsg={setErrorMsg} />
                                </FlexBox>
                                <FlexBox style={{ flexGrow: "1" }}></FlexBox>
                            </ContentPanelBody>
                        </ContentPanel>
                        <MirrorInfoPanel info={info} updateSettings={updateSettings} namespace={namespace} style={{ width: "100%", height: "100%", flex: 1 }} />
                    </FlexBox>
                    <ContentPanel style={{ width: "100%", minHeight: "15vh", flex: 1 }}>
                        <ContentPanelTitle>
                            <ContentPanelTitleIcon>
                                <VscAdd />
                            </ContentPanelTitleIcon>
                            <FlexBox className="gap" style={{ alignItems: "center" }}>Activity Logs</FlexBox>
                        </ContentPanelTitle>
                        <ContentPanelBody>
                            <ActivityLogs activity={activity} namespace={namespace} setErrorMsg={setErrorMsg} />
                        </ContentPanelBody>
                    </ContentPanel>

                </FlexBox>
            </Loader>
        </>
    );
}

export function MirrorReadOnlyBadge(props) {
    return (
        <Tippy content={`This mirrors contents are currently read-only. This can be unlocked in mirror setttings`} trigger={'mouseenter focus'} zIndex={10}>
            <div>
                <Button className={`cancel-label small disabled-no-filter shadow`} style={{ fontWeight: "bold", width: "fit-content", whiteSpace: "nowrap"}}>
                    <FlexBox className="row center gap-sm">
                        <VscLock />ReadOnly
                    </FlexBox>
                </Button>
            </div>
        </Tippy>
    )
}

export function MirrorWritableBadge(props) {
    return (
        <Tippy content={`This mirrors contents are currently writable. This can be unlocked in mirror setttings`} trigger={'mouseenter focus'} zIndex={10}>
            <div>
                <Button className={`running-label small disabled-no-filter shadow`} style={{ fontWeight: "bold", width: "fit-content", whiteSpace: "nowrap"}}>
                    <FlexBox className="row center gap-sm">
                        <VscUnlock />Writable
                    </FlexBox>
                </Button>
            </div>
        </Tippy>
    )
}