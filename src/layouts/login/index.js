import { useState } from "react"
import FlexBox from '../../components/flexbox';
import Logo from '../../assets/nav-logo.png'
import Button from "../../components/button";

export default function Login(props) {
    const {setLogin, setAKey} = props

    const [apiKey, setApiKey] = useState("")

    async function setAuth() {
        localStorage.setItem('apikey', apiKey)
        setAKey(apiKey)
        setApiKey("")
        setLogin(false)
    }

    return (
        <div style={{height:"100vh"}}>
            <FlexBox style={{height:"90%"}}>
                <FlexBox className="col gap tall" style={{ gap: "12px" }}>
                    <FlexBox className="navbar-logo" style={{margin:0}}>
                        <img alt="logo" src={Logo} />
                    </FlexBox>
                    <FlexBox className="col gap" style={{alignItems:"center", justifyContent:'center'}}>
                        <FlexBox className="col gap" style={{padding:"15px", width:"400px", maxHeight:"150px", background:"white"}}>
                            <h1 style={{fontSize:"12pt"}}>Sign In</h1>
                            <div style={{paddingRight:"10px"}}>
                                <input type="password" value={apiKey} onChange={e=>setApiKey(e.target.value)} placeholder="Enter an apikey..."/> 
                            </div>
                            <div style={{display:"flex", justifyContent:"flex-end"}}>
                                <Button onClick={setAuth} className="small">
                                    Login
                                </Button>
                            </div>
                        </FlexBox>
                        <div>
                            {process.env.REACT_APP_VERSION} 
                        </div>
                    </FlexBox>
                </FlexBox>
            </FlexBox>
        </div>
    )
}