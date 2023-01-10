import { useCallback, useEffect, useRef, useState } from 'react';
import './style.css';

import Tippy from '@tippyjs/react';
import { VscAdd, VscCloudUpload, VscTerminal } from 'react-icons/vsc';
import TextareaAutosize from 'react-textarea-autosize';
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import FlexBox from '../../components/flexbox';
import HelpIcon from '../../components/help';
import HideShowButton from '../../components/hide-show';
import Modal  from '../../components/modal';
import { ClientFileUpload } from '../../components/navbar';


export const mirrorSettingInfoMetaInfo = {
    "url": {plainName: "URL", required: true, placeholder: "Enter repository URL", info: `URL to repository. If authentication method is SSH Key a git url must be used e.g. "git@github.com:direktiv/apps-svc.git". All other authentication methods must use HTTP/S urls`},
    "ref": {plainName: "Ref", required: true, placeholder: `Enter repository ref e.g. "main"`, info: `Repository reference to sync from. For example this could be a commit hash ("b139f0e") or branch ("main").`},
    "cron": {plainName: "Cron", required: false, placeholder: `Enter cron e.g. "0 * * * *"`, info: `Cron schedule expression for auto-syncing with remote repository. Example auto-sync every hour "0 * * * *". (Optional)`},
    "passphrase": {plainName: "Passphrase", required: false, placeholder: `Enter passphrase`, info: `Passphrase to decrypt keys. (Optional)`},
    "token": {plainName: "Token", required: true, placeholder: `Enter personal access token`, info: `Personal access token to used for authentication`},
    "publicKey": {plainName: "Public Key", required: true, placeholder: `Enter Public Key`, info: `Public SSH Key used for authenticating with repository.`},
    "privateKey": {plainName: "Private Key", required: true, placeholder: `Enter Private Key`, info: `Private SSH Key used for authenticating with repository.`},
}


export default function MirrorInfoPanel(props) {
    const { info, style, updateSettings } = props

    const [mirrorAuthMethod, setMirrorAuthMethod] = useState("none")
    const [mirrorAuthMethodOld, setMirrorAuthMethodOld] = useState("none")
    const [mirrorSettingsValid, setMirrorSettingsValid] = useState(false)
    const [mirrorSettingsValidateMsg, setMirrorSettingsValidateMsg] = useState("")

    const [showPassphrase, setShowPassphrase] = useState(false)

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
        "authMethod": false
    })

    const [mirrorErrors, setMirrorErrors] = useState({
        "publicKey": null,
        "privateKey": null,
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
            if (!infoChangesTrackerRef?.current?.url) {
                setInfoURL(info.url)
            }
        }

        if (info?.ref !== null) {
            setInfoRefOld(info.ref)
            if (!infoChangesTrackerRef?.current?.ref) {
                setInfoRef(info.ref)
            }
        }

        if (info?.cron !== null) {
            setInfoCronOld(info.cron)
            if (!infoChangesTrackerRef?.current?.cron) {
                setInfoCron(info.cron)
            }
        }

        if (!infoChangesTrackerRef?.current || !typeof infoChangesTrackerRef?.current === 'object'){
            return
        }

        let authMethod = "none"
        if (info.privateKey === "-" || (info.publicKey && info.publicKey !== "")) {
            // Auth method is SSH
            authMethod = "ssh"
        } else if (info.passphrase && info.privateKey === "" && info.publicKey === "") {
            // Auth method is token
            authMethod = "token"
        }

        if (!infoPendingChanges){
            setMirrorAuthMethod(authMethod)
        }

        setMirrorAuthMethodOld(authMethod)

    }, [info, infoPendingChanges])

    // Validate if form values is valid
    useEffect(() => {
        if (!infoPendingChanges){
            return
        }

        let authMethodChanged = mirrorAuthMethod !== mirrorAuthMethodOld
        let validateMsg = ""

        switch (mirrorAuthMethod) {
            case "ssh":
                // validate ssh
                if (authMethodChanged && infoPrivateKey === ""){
                    // Rule: Check if privateKey is set in form when switching to token auth method
                    validateMsg += "Private Key must be set. "
                } else if (infoPrivateKey === "" && info.privateKey !== "-"){
                    // Rule:  Check if privateKey is set in form or remotely
                    validateMsg += "Private Key must be set. "
                }

                if (authMethodChanged && infoPublicKey === ""){
                    // Rule: Check if publicKey is set in form when switching to token auth method
                    validateMsg += "Public Key must be set. "
                } else if (infoPublicKey === "" && info.publicKey.length === 0){
                    // Rule:  Check if publicKey is set in form or remotely
                    validateMsg += "Public Key must be set. "
                }

                break;
            case "token":
                //valid token
                if (!infoURL.startsWith("http")) {
                    validateMsg += "URL must use the http/s protocol. "
                }

                if (authMethodChanged && infoPassphrase === ""){
                    // Rule: Check if token is set in form when switching to token auth method
                    validateMsg += "Token must be set. "
                } else if (infoPassphrase === "" && info.passphrase !== "-"){
                    // Rule:  Check if token is set in form or remotely
                    validateMsg += "Token must be set. "
                }
                break;
            default:
                break;
        }

        if (validateMsg === "") {
            setMirrorSettingsValid(true)
        } else {
            setMirrorSettingsValid(false)
        }
        setMirrorSettingsValidateMsg(validateMsg)

        
    }, [info, infoPublicKey, infoPassphrase, infoURL, infoPrivateKey, infoPendingChanges, mirrorAuthMethod, mirrorAuthMethodOld])

    useEffect(() => {
        setInfoPendingChanges(infoChangesTracker.url || infoChangesTracker.ref || infoChangesTracker.cron || infoChangesTracker.passphrase || infoChangesTracker.publicKey || infoChangesTracker.privateKey || infoChangesTracker.authMethod)
    }, [infoChangesTracker])

    const passphrasePlaceholder = useCallback(() => {
        let placeholder = mirrorSettingInfoMetaInfo["passphrase"].placeholder
        if (!infoChangesTracker.passphrase && info?.passphrase  === "-" && mirrorAuthMethodOld === "ssh") {
            placeholder  = `●●●●●●●●●`
        }

        if (infoChangesTracker.passphrase && infoPassphrase === "") {
            placeholder = `DELETE PASSPHRASE`
        }

        return placeholder
    }, [info, mirrorAuthMethodOld, infoChangesTracker, infoPassphrase]);

    const publicKeyPlaceholder = useCallback(() => {
        let placeholder = mirrorSettingInfoMetaInfo["publicKey"].placeholder
        if (!infoChangesTracker.publicKey && info?.publicKey  !== "" && mirrorAuthMethodOld === "ssh") {
            placeholder  = `●●●●●●●●●`
        }

        if (infoChangesTracker.publicKey && infoPublicKey === "") {
            placeholder = `DELETE PUBLIC KEY`
        }

        return placeholder
    }, [info, mirrorAuthMethodOld, infoChangesTracker, infoPublicKey]);

    const privateKeyPlaceholder = useCallback(() => {
        let placeholder = mirrorSettingInfoMetaInfo["privateKey"].placeholder
        if (!infoChangesTracker.privateKey && info?.privateKey  !== "" && mirrorAuthMethodOld === "ssh") {
            placeholder  = `●●●●●●●●●`
        }

        if (infoChangesTracker.privateKey && infoPrivateKey === "") {
            placeholder = `DELETE PRIVATE KEY`
        }

        return placeholder
    }, [info, mirrorAuthMethodOld, infoChangesTracker, infoPrivateKey]);

    const accessTokenPlaceholder = useCallback(() => {
        let placeholder = mirrorSettingInfoMetaInfo["token"].placeholder
        if (!infoChangesTracker.passphrase && info?.token  !== "" && mirrorAuthMethodOld === "token") {
            placeholder  = `●●●●●●●●●`
        }

        if (infoChangesTracker.passphrase && infoPassphrase === "") {
            placeholder = `DELETE TOKEN`
        }

        return placeholder
    }, [info, mirrorAuthMethodOld, infoChangesTracker, infoPassphrase]);

    return (
        <ContentPanel id={`panel-mirror-info`} style={{ ...style }}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <VscAdd />
                </ContentPanelTitleIcon>
                <FlexBox gap style={{ alignItems: "center" }}><span className="hide-600">Mirror</span> Info
                    <FlexBox style={{ flex: "auto", justifyContent: "right", paddingRight: "6px", alignItems: "unset" }}>
                        <Tippy content={mirrorSettingsValidateMsg} disabled={mirrorSettingsValidateMsg === ""} trigger={'mouseenter focus'} zIndex={10}>
                            <div>                            
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
                                        <>
                                            Update <span className="hide-600">Settings</span>
                                        </>
                                    )}
                                    buttonProps={{
                                        disabled: !infoPendingChanges || !mirrorSettingsValid
                                    }}
                                    actionButtons={[
                                        {
                                            label: "Update Settings",

                                            onClick: async () => {
                                                let newSettings = {
                                                    "url": infoChangesTracker.url ? infoURL : "-",
                                                    "ref": infoChangesTracker.ref ? infoRef : "-",
                                                    "cron": infoChangesTracker.cron ? infoCron : "-",
                                                    "passphrase": infoChangesTracker.passphrase ? infoPassphrase : "-",
                                                    "publicKey": infoChangesTracker.publicKey ? infoPublicKey : "-",
                                                    "privateKey": infoChangesTracker.privateKey ? infoPrivateKey : "-",
                                                }

                                                if (mirrorAuthMethod === "token") {
                                                    newSettings["privateKey"] = ""
                                                    newSettings["publicKey"] = ""
                                                } else if (mirrorAuthMethod === "none") {
                                                    newSettings["passphrase"] = ""
                                                    newSettings["privateKey"] = ""
                                                    newSettings["publicKey"] = ""
                                                } else if (mirrorAuthMethod === "ssh" && !infoChangesTracker.passphrase) {
                                                    newSettings["passphrase"] = ""
                                                }

                                                await updateSettings(newSettings)

                                                resetStates()
                                            },

                                            buttonProps: {variant: "contained", color: "primary"},
                                            errFunc: () => { },
                                            closesModal: true
                                        },
                                        {
                                            label: "Cancel",
                                            onClick: () => { },
                                            buttonProps: {},
                                            errFunc: () => { },
                                            closesModal: true
                                        }
                                    ]}
                                >
                                    <FlexBox col gap style={{ height: "fit-content" }}>
                                        <FlexBox className="col center info-update-label">
                                            The following changes will been made
                                        </FlexBox>
                                        {
                                            infoChangesTracker.authMethod ?
                                            <FlexBox className="col center info-update-label" style={{textAlign: "center"}}>
                                                Warning: 
                                                { mirrorAuthMethod === "none" ? ` Changing authentication method to "None" will remove Passphrase, Public Key and Private Key from settings.`: ""}
                                                { mirrorAuthMethod === "token" ? ` Changing authentication method to "Access Token" will remove Public Key and Private Key from settings.`: ""}
                                                { mirrorAuthMethod === "ssh" ? ` Changing authentication method to "SSH Keys" will remove Access Token from settings.`: ""}
                                            </FlexBox>
                                            :
                                            <></>
                                        }
                                        {infoChangesTracker.url ?
                                            <FlexBox col gap style={{ paddingRight: "10px" }}>
                                                <FlexBox col gap="sm" center="x"  style={{}}>
                                                    <span className={`input-title readonly`} >URL</span>
                                                    { infoURL === "" ? <span className={`input-description readonly`}> Warning: URL will be deleted</span> : <></>}
                                                </FlexBox>
                                                <input className={`info-input-value readonly`} value={infoURL} />
                                            </FlexBox> : <></>}
                                        {infoChangesTracker.ref ?
                                            <FlexBox col gap style={{ paddingRight: "10px" }}>
                                                <FlexBox col gap="sm" center="x"  style={{}}>
                                                    <span className={`input-title readonly`} >Ref</span>
                                                    { infoRef === "" ? <span className={`input-description readonly`}> Warning: Ref will be deleted</span> : <></>}
                                                </FlexBox>
                                                <input className={`info-input-value readonly`} value={infoRef} />
                                            </FlexBox> : <></>}
                                        {infoChangesTracker.cron ?
                                            <FlexBox col gap style={{ paddingRight: "10px" }}>
                                                <FlexBox col gap="sm" center="x"  style={{}}>
                                                    <span className={`input-title readonly`} >Cron</span>
                                                    { infoCron === "" ? <span className={`input-description readonly`}> Warning: Cron will be deleted</span> : <></>}
                                                </FlexBox>
                                                <input className={`info-input-value readonly`} readonly={true} value={infoCron} />
                                            </FlexBox> : <></>}
                                        {infoChangesTracker.passphrase && mirrorAuthMethod === "token" ?
                                            <FlexBox col gap style={{ paddingRight: "10px" }}>
                                                <FlexBox col gap="sm" center="x"  style={{}}>
                                                    <span className={`input-title readonly`} >Token</span>
                                                    { infoPassphrase === "" ? <span className={`input-description readonly`}> Warning: Token will be deleted</span> : <></>}
                                                </FlexBox>
                                                <textarea className={`info-textarea-value readonly`} readonly={true} rows={5} style={{ width: "100%", resize: "none" }} value={infoPassphrase} />
                                            </FlexBox> : <></>}
                                        {infoChangesTracker.passphrase && mirrorAuthMethod === "ssh" ?
                                            <FlexBox col gap style={{ paddingRight: "10px" }}>
                                                <FlexBox col gap="sm" center="x"  style={{}}>
                                                    <span className={`input-title readonly`} >Passphrase</span>
                                                    { infoPassphrase === "" ? <span className={`input-description readonly`}> Warning: Passphrase will be deleted</span> : <></>}
                                                </FlexBox>
                                                <input className={`info-input-value readonly`} readonly={true} type="password" value={infoPassphrase} />
                                            </FlexBox> : <></>}
                                        {infoChangesTracker.publicKey && mirrorAuthMethod === "ssh" ?
                                            <FlexBox col gap style={{ paddingRight: "10px" }}>
                                                <FlexBox col gap="sm" center="x"  style={{}}>
                                                    <span className={`input-title readonly`} >Public Key</span>
                                                    { infoPublicKey === "" ? <span className={`input-description readonly`}> Warning: Public Key will be deleted</span> : <></>}
                                                </FlexBox>
                                                <textarea className={`info-textarea-value readonly`} readonly={true} rows={5} style={{ width: "100%", resize: "none" }} value={infoPublicKey} />
                                            </FlexBox> : <></>}
                                        {infoChangesTracker.privateKey && mirrorAuthMethod === "ssh" ?
                                            <FlexBox col gap style={{ paddingRight: "10px" }}>
                                                <FlexBox col gap="sm" center="x"  style={{}}>
                                                    <span className={`input-title readonly`} >Private Key</span>
                                                    { infoPrivateKey === "" ? <span className={`input-description readonly`}> Warning: Private Key will be deleted</span> : <></>}
                                                </FlexBox>
                                                <textarea className={`info-textarea-value readonly`} readonly={true} rows={5} style={{ width: "100%", resize: "none" }} value={infoPrivateKey} />
                                            </FlexBox> : <></>}
                                    </FlexBox>
                            </Modal>
                            </div>
                        </Tippy>
                    </FlexBox>
                </FlexBox>
            </ContentPanelTitle>
            <ContentPanelBody style={{ overflow: "auto" }}>
                <FlexBox col gap style={{ height: "fit-content" }}>
                    <FlexBox col gap="sm" style={{ }}>
                        <FlexBox row gap="sm" style={{ justifyContent: "flex-start" }}>
                            <span className={`input-title`}>Authentication Method</span>
                        </FlexBox>
                        <div style={{ width: "100%", paddingRight: "12px", display: "flex" }}>
                            <select style={{ width: "100%" }} defaultValue={mirrorAuthMethod} value={mirrorAuthMethod} onChange={(e) => {
                                    // Track authmethod is changed 
                                    const newAuthMethod = e.target.value
                                    setMirrorAuthMethod(newAuthMethod)
                                    setInfoChangesTracker((old) => {
                                        old.authMethod = newAuthMethod !== mirrorAuthMethodOld
                                        return { ...old }
                                    })
                                }}>
                                <option value="none">None</option>
                                <option value="ssh">SSH Keys</option>
                                <option value="token">Access Token</option>
                            </select>
                        </div>
                    </FlexBox>
                    <FlexBox col gap="sm" style={{ paddingRight: "10px" }}>
                        <FlexBox row style={{ justifyContent: "space-between" }}>
                            <FlexBox row gap="sm" style={{ justifyContent: "flex-start" }}>
                                <span className={`input-title ${infoChangesTracker.url ? "edited" : ""}`}>URL</span>
                                <HelpIcon msg={mirrorSettingInfoMetaInfo["url"].info} />
                            </FlexBox>
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
                        }} placeholder={mirrorSettingInfoMetaInfo["url"].placeholder} />
                    </FlexBox>
                    <FlexBox col gap="sm" style={{ paddingRight: "10px" }}>
                        <FlexBox row style={{ justifyContent: "space-between" }}>
                            <FlexBox row gap="sm" style={{ justifyContent: "flex-start" }}>
                                <span className={`input-title ${infoChangesTracker.ref ? "edited" : ""}`}>Ref</span>
                                <HelpIcon msg={mirrorSettingInfoMetaInfo["ref"].info} />
                            </FlexBox>
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
                        }} placeholder={mirrorSettingInfoMetaInfo["ref"].placeholder} />
                    </FlexBox>
                    <FlexBox col gap="sm" style={{ paddingRight: "10px" }}>
                        <FlexBox row style={{ justifyContent: "space-between" }}>
                            <FlexBox row gap="sm" style={{ justifyContent: "flex-start" }}>
                                <span className={`input-title ${infoChangesTracker.cron ? "edited" : ""}`}>Cron</span>
                                <HelpIcon msg={mirrorSettingInfoMetaInfo["cron"].info} />
                            </FlexBox>
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
                        }} placeholder={mirrorSettingInfoMetaInfo["cron"].placeholder} />
                    </FlexBox>

                    {/* PERSONAL ACCESS TOKEN Auth Inputs */}
                    {
                        mirrorAuthMethod === "token" ?
                            <FlexBox col gap="sm" style={{ paddingRight: "10px" }}>
                                <FlexBox row style={{ justifyContent: "space-between" }}>
                                    <FlexBox row gap="sm" style={{ justifyContent: "flex-start" }}>
                                        <span className={`input-title ${infoChangesTracker.passphrase ? "edited" : ""}`}>Token</span>
                                        <HelpIcon msg={mirrorSettingInfoMetaInfo["token"].info} />
                                    </FlexBox>
                                    <span className={`info-input-undo ${infoChangesTracker.passphrase ? "" : "hide"}`} onClick={(e) => {
                                        setInfoPassphrase(infoPassphraseOld)
                                        setInfoChangesTracker((old) => {
                                            old.passphrase = false
                                            return { ...old }
                                        })
                                    }}>Undo Changes</span>
                                </FlexBox>
                                <TextareaAutosize style={{ width: "100%", resize: "none", padding: "11px 1px 11px 8px" }} minRows={infoChangesTracker.passphrase || info?.passphrase === "" ? 2 : 1} maxRows={5} value={infoPassphrase} onChange={(e) => {
                                    setInfoPassphrase(e.target.value)
                                    setInfoChangesTracker((old) => {
                                        old.passphrase = true
                                        return { ...old }
                                    })
                                }} placeholder={accessTokenPlaceholder()} />
                            </FlexBox> : <></>
                    }

                    {/* SSH Auth Inputs */}
                    {
                        mirrorAuthMethod === "ssh" ?
                            <>
                                <FlexBox col gap="sm" style={{ paddingRight: "10px" }}>
                                    <FlexBox row style={{ justifyContent: "space-between" }}>
                                        <FlexBox row gap="sm" style={{ justifyContent: "flex-start" }}>
                                            <span className={`input-title ${infoChangesTracker.passphrase ? "edited" : ""}`}>Passphrase</span>
                                            <HelpIcon msg={mirrorSettingInfoMetaInfo["passphrase"].info} />
                                            <HideShowButton show={showPassphrase} setShow={setShowPassphrase} field={"Passphrase"}/>
                                        </FlexBox>
                                        <span className={`info-input-undo ${infoChangesTracker.passphrase ? "" : "hide"}`} onClick={(e) => {
                                            setInfoPassphrase(infoPassphraseOld)
                                            setInfoChangesTracker((old) => {
                                                old.passphrase = false
                                                return { ...old }
                                            })
                                        }}>Undo Changes</span>
                                    </FlexBox>
                                    <input type={`${showPassphrase ? "text" : "password"}`} value={infoPassphrase} onChange={(e) => {
                                        setInfoPassphrase(e.target.value)
                                        setInfoChangesTracker((old) => {
                                            old.passphrase = true
                                            return { ...old }
                                        })
                                    }} placeholder={passphrasePlaceholder()} />
                                </FlexBox>
                                <FlexBox col gap="sm" style={{ paddingRight: "10px" }}>
                                    <FlexBox row style={{ justifyContent: "space-between" }}>
                                        <FlexBox row gap="sm" style={{ justifyContent: "flex-start" }}>
                                            <span className={`input-title ${infoChangesTracker.publicKey ? "edited" : ""}`}>Public Key</span>
                                            <HelpIcon msg={mirrorSettingInfoMetaInfo["publicKey"].info} />
                                        </FlexBox>
                                        <FlexBox row gap style={{ justifyContent: "flex-end", gap: "12px" }}>
                                            <ClientFileUpload
                                                setFile={(fileData) => {
                                                    setInfoPublicKey(fileData)
                                                    setInfoChangesTracker((old) => {
                                                        old.publicKey = true
                                                        return { ...old }
                                                    })
                                                }}
                                                setError={(errorMsg) => {
                                                    let newErrors = mirrorErrors
                                                    newErrors["publicKey"] = errorMsg
                                                    setMirrorErrors({ ...newErrors })
                                                }}
                                                maxSize={40960}
                                            >
                                                <Tippy content={mirrorErrors["publicKey"] ? mirrorErrors["publicKey"] : `Upload key plaintext file content to ${mirrorSettingInfoMetaInfo["publicKey"].plainName} input. Warning: this will replace current ${mirrorSettingInfoMetaInfo["publicKey"].plainName} content`} trigger={'click mouseenter focus'} onHide={() => {
                                                    let newErrors = mirrorErrors
                                                    newErrors["publicKey"] = null
                                                    setMirrorErrors({ ...newErrors })
                                                }}
                                                    zIndex={10}>
                                                    <div className='input-title-button'>
                                                        <FlexBox center row gap="sm" style={{ justifyContent: "flex-end", marginRight: "-6px" }}>
                                                            <span onClick={(e) => {
                                                            }}>Upload</span>
                                                            <VscCloudUpload />
                                                        </FlexBox>
                                                    </div>
                                                </Tippy>
                                            </ClientFileUpload>
                                            <span className={`info-input-undo ${infoChangesTracker.publicKey ? "" : "hide"}`} onClick={(e) => {
                                                setInfoPublicKey(infoPublicKeyOld)
                                                setInfoChangesTracker((old) => {
                                                    old.publicKey = false
                                                    return { ...old }
                                                })
                                            }}>Undo Changes</span>
                                        </FlexBox>
                                    </FlexBox>
                                    <TextareaAutosize style={{ width: "100%", resize: "none", padding: "11px 1px 11px 8px" }} minRows={infoChangesTracker.publicKey || info?.publicKey === "" ? 2 : 1} maxRows={5} value={infoPublicKey} onChange={(e) => {
                                        setInfoPublicKey(e.target.value)
                                        setInfoChangesTracker((old) => {
                                            old.publicKey = true
                                            return { ...old }
                                        })
                                    }} placeholder={publicKeyPlaceholder()} />
                                </FlexBox>
                                <FlexBox col gap="sm" style={{ paddingRight: "10px" }}>
                                    <FlexBox row style={{ justifyContent: "space-between" }}>
                                        <FlexBox row gap="sm" style={{ justifyContent: "flex-start" }}>
                                            <span className={`input-title ${infoChangesTracker.privateKey ? "edited" : ""}`} >Private Key</span>
                                            <HelpIcon msg={mirrorSettingInfoMetaInfo["privateKey"].info} />
                                        </FlexBox>
                                        <FlexBox row gap style={{ justifyContent: "flex-end", gap: "12px" }}>
                                            <ClientFileUpload
                                                setFile={(fileData) => {
                                                    setInfoPrivateKey(fileData)
                                                    setInfoChangesTracker((old) => {
                                                        old.privateKey = true
                                                        return { ...old }
                                                    })
                                                }}
                                                setError={(errorMsg) => {
                                                    let newErrors = mirrorErrors
                                                    newErrors["privateKey"] = errorMsg
                                                    setMirrorErrors({ ...newErrors })
                                                }}
                                                maxSize={40960}
                                            >
                                                <Tippy content={mirrorErrors["privateKey"] ? mirrorErrors["privateKey"] : `Upload key plaintext file content to ${mirrorSettingInfoMetaInfo["privateKey"].plainName} input. Warning: this will replace current ${mirrorSettingInfoMetaInfo["privateKey"].plainName} content`} trigger={'click mouseenter focus'} onHide={() => {
                                                    let newErrors = mirrorErrors
                                                    newErrors["privateKey"] = null
                                                    setMirrorErrors({ ...newErrors })
                                                }}
                                                    zIndex={10}>
                                                    <div className='input-title-button'>
                                                        <FlexBox center row gap="sm" style={{ justifyContent: "flex-end", marginRight: "-6px" }}>
                                                            <span onClick={(e) => {
                                                            }}>Upload</span>
                                                            <VscCloudUpload />
                                                        </FlexBox>
                                                    </div>
                                                </Tippy>
                                            </ClientFileUpload>
                                            <span className={`info-input-undo ${infoChangesTracker.privateKey ? "" : "hide"}`} onClick={(e) => {
                                                setInfoPrivateKey(infoPrivateKeyOld)
                                                setInfoChangesTracker((old) => {
                                                    old.privateKey = false
                                                    return { ...old }
                                                })
                                            }}>Undo Changes</span>
                                        </FlexBox>
                                    </FlexBox>
                                    <TextareaAutosize style={{ width: "100%", resize: "none", padding: "11px 1px 11px 8px" }} minRows={infoChangesTracker.privateKey || info?.privateKey === "" ? 2 : 1} maxRows={5} value={infoPrivateKey} onChange={(e) => {
                                        setInfoPrivateKey(e.target.value)
                                        setInfoChangesTracker((old) => {
                                            old.privateKey = true
                                            return { ...old }
                                        })
                                    }} placeholder={privateKeyPlaceholder()} />
                                </FlexBox>
                            </> : <></>
                    }
                </FlexBox>

            </ContentPanelBody>
        </ContentPanel>

    );
}