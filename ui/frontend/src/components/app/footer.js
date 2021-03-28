import Vorteil from "../../img/logo.png"

export function Footer() {

    return (
        <div className="footer" style={{marginTop: "20px"}}>
            <div style={{width: "100%", textAlign: "center", color: "white"}}>
                Powered by <a href="https://vorteil.io" className="footer-link"><img style={{height: "30px"}}
                                                                                     alt="main-company" src={Vorteil}/>Vorteil.io</a>
            </div>
        </div>
    )

}

export default Footer