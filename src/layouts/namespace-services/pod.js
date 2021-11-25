import { useParams } from "react-router"
import FlexBox from "../../components/flexbox"


export default function PodPanel(props) {
    const {namespace} = props
    const {service, revision} = useParams()
    if(!namespace) {
        return ""
    }

    return(
        <FlexBox className="gap wrap" style={{paddingRight:"8px"}}>
            pod page
        </FlexBox>
    )
}