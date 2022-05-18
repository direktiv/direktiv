import React, { useEffect, useState, useCallback, useRef } from 'react';
import './style.css';

import ContentPanel, { ContentPanelBody, ContentPanelHeaderButton, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import FlexBox from '../../components/flexbox';
import { VscAdd, VscTerminal } from 'react-icons/vsc';
import Modal, { ButtonDefinition } from '../../components/modal';

export const mirrorSettingInfoMetaInfo = {
    "url": {plainName: "URL", placeholder: "", info: null},
    "ref": {plainName: "Ref", placeholder: "", info: null},
    "cron": {plainName: "Cron", placeholder: "", info: null},
    "passphrase": {plainName: "Passphrase", placeholder: "", info: null},
    "publicKey": {plainName: "Public Key", placeholder: "", info: null},
    "privateKey": {plainName: "Private Key", placeholder: "", info: null},
}


export default function MirrorInfoPanel(props) {
    const { info, style, updateSettings } = props

    // Mirror Info States
    const [infoURL, setInfoURL] = useState("")
    const [infoRef, setInfoRef] = useState("")
    const [infoCron, setInfoCron] = useState("")
    const [infoPublicKey, setInfoPublicKey] = useState("")
    const [infoPrivateKey, setInfoPrivateKey] = useState("")
    const [infoPassphrase, setInfoPassphrase] = useState("")

    const [infoURLOld, setInfoURLOld] = useState("")
    const [infoRefOld, setInfoRefOld] = useState("")
    const [infoCronOld, setInfoCronOld] = useState("")
    const [infoPublicKeyOld,] = useState("")
    const [infoPrivateKeyOld,] = useState("")
    const [infoPassphraseOld,] = useState("")

    const [infoPendingChanges, setInfoPendingChanges] = useState(false)
    const [infoChangesTracker, setInfoChangesTracker] = useState({
        "url": false,
        "ref": false,
        "cron": false,
        "passphrase": false,
        "publicKey": false,
        "privateKey": false,
    })

    const infoChangesTrackerRef = useRef(infoChangesTracker)

    const resetStates = useCallback(() => {
        setInfoChangesTracker({
            "url": false,
            "ref": false,
            "cron": false,
            "passphrase": false,
            "publicKey": false,
            "privateKey": false,
        })

        setInfoURL(infoURLOld)
        setInfoRef(infoRefOld)
        setInfoCron(infoCronOld)
        setInfoPublicKey(infoPublicKeyOld)
        setInfoPrivateKey(infoPrivateKeyOld)
        setInfoPassphrase(infoPassphraseOld)
    }, [infoURLOld, infoRefOld, infoCronOld, infoPublicKeyOld, infoPrivateKeyOld, infoPassphraseOld])

    useEffect(() => {
        infoChangesTrackerRef.current = infoChangesTracker
    }, [infoChangesTracker])



    useEffect(() => {
        if (!info) {
            return
        }

        if (info?.url !== null) {
            setInfoURLOld(info.url)
            if (!infoChangesTrackerRef.url) {
                setInfoURL(info.url)
            }
        }

        if (info?.ref !== null) {
            setInfoRefOld(info.ref)
            if (!infoChangesTrackerRef.ref) {
                setInfoRef(info.ref)
            }
        }

        if (info?.cron !== null) {
            setInfoCronOld(info.cron)
            if (!infoChangesTrackerRef.cron) {
                setInfoCron(info.cron)
            }
        }

    }, [info])

    useEffect(() => {
        setInfoPendingChanges(infoChangesTracker.url || infoChangesTracker.ref || infoChangesTracker.cron || infoChangesTracker.passphrase || infoChangesTracker.publicKey || infoChangesTracker.privateKey)
    }, [infoChangesTracker])

    return (
        <ContentPanel id={`panel-mirror-info`} style={{ ...style }}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <VscAdd />
                </ContentPanelTitleIcon>
                <FlexBox className="gap" style={{ alignItems: "center" }}>Mirror Info
                    <FlexBox style={{ flex: "auto", justifyContent: "right", paddingRight: "6px", alignItems: "unset" }}>
                        <ContentPanelHeaderButton className={`${infoPendingChanges ? "" : "disabled"}`} style={infoPendingChanges ? {} : { color: "grey" }}>
                            <Modal
                                escapeToCancel
                                activeOverlay
                                title="Update Mirror Settings"
                                titleIcon={
                                    <VscTerminal />
                                }
                                style={{
                                    maxWidth: "260px"
                                }}
                                modalStyle={{
                                    width: "300px"
                                }}
                                button={(
                                    <div>
                                        Update Settings
                                    </div>
                                )}
                                actionButtons={[
                                    ButtonDefinition("Update Settings", async () => {
                                        await updateSettings({
                                            "url": infoChangesTracker.url ? infoURL : "-",
                                            "ref": infoChangesTracker.ref ? infoRef : "-",
                                            "cron": infoChangesTracker.cron ? infoCron : "-",
                                            "passphrase": infoChangesTracker.passphrase ? infoPassphrase : "-",
                                            "publicKey": infoChangesTracker.publicKey ? infoPublicKey : "-",
                                            "privateKey": infoChangesTracker.privateKey ? infoPrivateKey : "-",

                                        })

                                        resetStates()
                                    }, "small blue", () => { }, true, false),
                                    ButtonDefinition("Cancel", () => { }, "small light", () => { }, true, false)
                                ]}
                            >
                                <FlexBox className="col gap" style={{ height: "fit-content" }}>
                                    <FlexBox className="col center info-update-label">
                                        The following changes will been made
                                    </FlexBox>
                                    {infoChangesTracker.url ?
                                        <FlexBox className="col gap" style={{ paddingRight: "10px" }}>
                                            <span className={`input-title readonly`}>URL</span>
                                            <input className={`info-input-value readonly`} value={infoURL} />
                                        </FlexBox> : <></>}
                                    {infoChangesTracker.ref ?
                                        <FlexBox className="col gap" style={{ paddingRight: "10px" }}>
                                            <span className={`input-title readonly`}>Ref</span>
                                            <input className={`info-input-value readonly`} value={infoRef} />
                                        </FlexBox> : <></>}
                                    {infoChangesTracker.cron ?
                                        <FlexBox className="col gap" style={{ paddingRight: "10px" }}>
                                            <span className={`input-title readonly`}>Cron</span>
                                            <input className={`info-input-value readonly`} readonly={true} value={infoCron} />
                                        </FlexBox> : <></>}
                                    {infoChangesTracker.passphrase ?
                                        <FlexBox className="col gap" style={{ paddingRight: "10px" }}>
                                            <span className={`input-title readonly`}>Passphrase</span>
                                            <input className={`info-input-value readonly`} readonly={true} type="password" value={infoPassphrase} />
                                        </FlexBox> : <></>}
                                    {infoChangesTracker.publicKey ?
                                        <FlexBox className="col gap" style={{ paddingRight: "10px" }}>
                                            <span className={`input-title readonly`}>Public Key</span>
                                            <textarea className={`info-textarea-value readonly`} readonly={true} style={{ width: "100%", resize: "none" }} value={infoPublicKey} />
                                        </FlexBox> : <></>}
                                    {infoChangesTracker.privateKey ?
                                        <FlexBox className="col gap" style={{ paddingRight: "10px" }}>
                                            <span className={`input-title readonly`} >Private Key</span>
                                            <textarea className={`info-textarea-value readonly`} readonly={true} style={{ width: "100%", resize: "none" }} value={infoPrivateKey} />
                                        </FlexBox> : <></>}
                                </FlexBox>
                            </Modal>
                        </ContentPanelHeaderButton>
                    </FlexBox>
                </FlexBox>
            </ContentPanelTitle>
            <ContentPanelBody style={{ overflow: "auto" }}>
                <FlexBox className="col gap" style={{ height: "fit-content" }}>
                    <FlexBox className="col gap-md" style={{ paddingRight: "10px" }}>
                        <FlexBox className="row" style={{ justifyContent: "space-between" }}>
                            <span className={`input-title ${infoChangesTracker.url ? "edited" : ""}`}>URL</span>
                            <span className={`info-input-undo ${infoChangesTracker.url ? "" : "hide"}`} onClick={(e) => {
                                setInfoURL(infoURLOld)
                                setInfoChangesTracker((old) => {
                                    old.url = false
                                    return { ...old }
                                })
                            }}>Undo Changes</span>
                        </FlexBox>
                        <input value={infoURL} onChange={(e) => {
                            setInfoURL(e.target.value)
                            setInfoChangesTracker((old) => {
                                old.url = true
                                return { ...old }
                            })
                        }} placeholder="Enter URL" />
                    </FlexBox>
                    <FlexBox className="col gap-md" style={{ paddingRight: "10px" }}>
                        <FlexBox className="row" style={{ justifyContent: "space-between" }}>
                            <span className={`input-title ${infoChangesTracker.ref ? "edited" : ""}`}>Ref</span>
                            <span className={`info-input-undo ${infoChangesTracker.ref ? "" : "hide"}`} onClick={(e) => {
                                setInfoRef(infoRefOld)
                                setInfoChangesTracker((old) => {
                                    old.ref = false
                                    return { ...old }
                                })
                            }}>Undo Changes</span>
                        </FlexBox>
                        <input value={infoRef} onChange={(e) => {
                            setInfoRef(e.target.value)
                            setInfoChangesTracker((old) => {
                                old.ref = true
                                return { ...old }
                            })
                        }} placeholder="Enter Ref" />
                    </FlexBox>
                    <FlexBox className="col gap-md" style={{ paddingRight: "10px" }}>
                        <FlexBox className="row" style={{ justifyContent: "space-between" }}>
                            <span className={`input-title ${infoChangesTracker.cron ? "edited" : ""}`}>Cron</span>
                            <span className={`info-input-undo ${infoChangesTracker.cron ? "" : "hide"}`} onClick={(e) => {
                                setInfoCron(infoCronOld)
                                setInfoChangesTracker((old) => {
                                    old.cron = false
                                    return { ...old }
                                })
                            }}>Undo Changes</span>
                        </FlexBox>
                        <input value={infoCron} onChange={(e) => {
                            setInfoCron(e.target.value)
                            setInfoChangesTracker((old) => {
                                old.cron = true
                                return { ...old }
                            })
                        }} placeholder="Enter cron" />
                    </FlexBox>
                    <FlexBox className="col gap-md" style={{ paddingRight: "10px" }}>
                        <FlexBox className="row" style={{ justifyContent: "space-between" }}>
                            <span className={`input-title ${infoChangesTracker.passphrase ? "edited" : ""}`}>Passphrase</span>
                            <span className={`info-input-undo ${infoChangesTracker.passphrase ? "" : "hide"}`} onClick={(e) => {
                                setInfoPassphrase(infoPassphraseOld)
                                setInfoChangesTracker((old) => {
                                    old.passphrase = false
                                    return { ...old }
                                })
                            }}>Undo Changes</span>
                        </FlexBox>
                        <input type="password" value={infoPassphrase} onChange={(e) => {
                            setInfoPassphrase(e.target.value)
                            setInfoChangesTracker((old) => {
                                old.passphrase = true
                                return { ...old }
                            })
                        }} placeholder="Enter Passphrase" />
                    </FlexBox>
                    <FlexBox className="col gap-md" style={{ paddingRight: "10px" }}>
                        <FlexBox className="row" style={{ justifyContent: "space-between" }}>
                            <span className={`input-title ${infoChangesTracker.publicKey ? "edited" : ""}`}>Public Key</span>
                            <span className={`info-input-undo ${infoChangesTracker.publicKey ? "" : "hide"}`} onClick={(e) => {
                                setInfoPublicKey(infoPublicKeyOld)
                                setInfoChangesTracker((old) => {
                                    old.publicKey = false
                                    return { ...old }
                                })
                            }}>Undo Changes</span>
                        </FlexBox>
                        <textarea style={{ width: "100%", resize: "none" }} rows={5} value={infoPublicKey} onChange={(e) => {
                            setInfoPublicKey(e.target.value)
                            setInfoChangesTracker((old) => {
                                old.publicKey = true
                                return { ...old }
                            })
                        }} placeholder="Enter Public Key" />
                    </FlexBox>
                    <FlexBox className="col gap-md" style={{ paddingRight: "10px" }}>
                        <FlexBox className="row" style={{ justifyContent: "space-between" }}>
                            <span className={`input-title ${infoChangesTracker.privateKey ? "edited" : ""}`} >Private Key</span>
                            <span className={`info-input-undo ${infoChangesTracker.privateKey ? "" : "hide"}`} onClick={(e) => {
                                setInfoPrivateKey(infoPrivateKeyOld)
                                setInfoChangesTracker((old) => {
                                    old.privateKey = false
                                    return { ...old }
                                })
                            }}>Undo Changes</span>
                        </FlexBox>
                        <textarea type="password" style={{ width: "100%", resize: "none" }} rows={5} value={infoPrivateKey} onChange={(e) => {
                            setInfoPrivateKey(e.target.value)
                            setInfoChangesTracker((old) => {
                                old.privateKey = true
                                return { ...old }
                            })
                        }} placeholder="Enter Private Key" />
                    </FlexBox>
                </FlexBox>

            </ContentPanelBody>
        </ContentPanel>

    );
}