import {useHistory, useParams} from "react-router-dom";
import {Journal} from "react-bootstrap-icons"
import {Button, OverlayTrigger, Tooltip} from "react-bootstrap"

const renderTooltip = (props) => (
    <Tooltip id="button-tooltip" {...props}>
        Back to Workflow
    </Tooltip>
);


export function BackToWorkflow(props) {
    const params = useParams()
    const history = useHistory()
    return (
        <>
            <OverlayTrigger
                placement="left"
                delay={{show: 100, hide: 150}}
                overlay={renderTooltip}
            >
          <span>
            <Button
                style={{
                    marginLeft: "6px",
                    background: "#e9ecef",
                    borderColor: "#e0e0e0",
                    textAlign: "center",
                    padding: "0.375rem 0.375rem 0.375rem 0.375rem",
                }}
                variant="light"
                onClick={() => {
                    history.push(`/p/${params.namespace}/w/${params.workflow}`)

                }}
            >
              <Journal width="1.5em" height="1.5em"/>
            </Button>
          </span>
            </OverlayTrigger>
        </>
    )
}

export default BackToWorkflow