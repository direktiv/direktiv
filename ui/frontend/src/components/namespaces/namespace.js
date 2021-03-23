import React, {useCallback, useContext, useEffect, useMemo, useState} from 'react'
import '../../css/namespace.css'
import ReactPaginate from 'react-paginate'

import {useLocation, useParams} from 'react-router-dom'
import logoColor from "img/logo-color.png";

import {Alert, Container} from 'react-bootstrap'
import {XCircle} from 'react-bootstrap-icons'

import Row from 'react-bootstrap/Row'
import Col from 'react-bootstrap/Col'

import Fuse from 'fuse.js'

import NamespaceActions from 'components/namespaces/actions'
import WorkflowList from 'components/namespaces/workflows'
import {QueryParams} from 'util/params'
import ServerContext from 'components/app/context'


export default function Namespace(props) {
    const [workflows, setWorkflows] = useState([])
    const [permErr, setPermError] = useState("")
    const [workflowErr, setWorkflowError] = useState("")
    const [loader, setLoader] = useState(false)
    const [pagination, setPagination] = useState({total: 0, offset: 0})
    const [searchInfo, setSearchInfo] = useState({results: [], pattern: ""})
    const context = useContext(ServerContext);
    const fuse = useMemo(() => new Fuse(workflows, {
        threshold: 0.4, distance: 50, keys: [
            "id",
            "description"
        ]
    }), [workflows]);

    let {namespace} = useParams()

    let params = QueryParams(useLocation().search)
    if (!params.p) {
        params.p = 1
    }
    if (!params.q) {
        params.q = ""
    }

    const clearError = () => {
        setPermError("")
        setWorkflowError("")
    }

    function updateWorkflows(newWorkflows) {
        setWorkflows(newWorkflows)
    }

    const fetchWorkflows = useCallback(() => {
        async function fetchWFs() {
            try {
                setLoader(true)
                let resp = await context.Fetch(`/namespaces/${namespace}/workflows?offset=${pagination.offset}`, {})
                if (!resp.ok) {
                    setLoader(false)
                    throw resp
                } else {
                    let json = await resp.json()
                    updateWorkflows(json.workflows)
                    setPagination({...pagination, total: json.total})
                }
            } catch (e) {
                setWorkflowError(e.message)
            }
            setLoader(false)
        }

        fetchWFs()
    }, [context.Fetch, namespace, pagination.offset])

    // Fetch data on mount
    useEffect(() => {
        fetchWorkflows()
    }, [fetchWorkflows, context.Fetch, namespace])

    useEffect(() => {
        if (params.q === searchInfo.pattern) {
            return
        }
        let results = []
        fuse.search(params.q).forEach(res => results.push(res.item))
        setSearchInfo({results: results, pattern: params.q})
    }, [params, searchInfo.pattern, fuse])

    function searchResults() {
        if (searchInfo.pattern === "") {
            return workflows
        } else {
            return searchInfo.results
        }
    }

    return (<>
            <Row style={{margin: "0px"}}>
                <Col style={{marginBottom: "15px"}} id="namespace-header">
                    <div className="namespace-actions">
                        <div id="namespace-actions-title" className="namespace-actions-box">
                            <h4>
                                {namespace}
                            </h4>
                        </div>
                        <div id="namespace-actions-options" className="namespace-actions-box">
                            <NamespaceActions namespace={namespace} q={params.q} p={params.p}
                                              onNew={() => {
                                                  fetchWorkflows() // fetchWorkflows again
                                              }}
                            />
                        </div>
                    </div>
                    <div className="padded-border"></div>
                </Col>
                <Col style={{marginBottom: "15px"}} xs={12}>
                    {workflowErr !== "" || permErr !== "" ?
                        <Row style={{margin: "0px"}}>
                            <Col style={{marginBottom: "15px"}} xs={12} md={12}>
                                <Alert variant="danger">
                                    <Container>
                                        <Row>
                                            <Col sm={11}>
                                                Fetching Workflows: {workflowErr !== "" ? workflowErr : permErr}
                                            </Col>
                                            <Col sm={1} style={{textAlign: "right", paddingRight: "0"}}>
                                                <XCircle style={{cursor: "pointer", fontSize: "large"}} onClick={() => {
                                                    clearError()
                                                }}/>
                                            </Col>
                                        </Row>
                                    </Container>
                                </Alert>
                            </Col>
                        </Row>
                        :
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
                                <>
                                    <div>
                                        <h5>
                                            Workflows
                                        </h5>
                                        <h4 style={{color: "#999999", fontSize: "1rem"}}>
                                            {(searchInfo.pattern !== "")
                                                ? (<>{searchInfo.results.length} Results{' '}</>)
                                                :
                                                (<></>)}
                                        </h4>
                                    </div>
                                    {searchResults().length > 0 ?

                                        <WorkflowList namespace={namespace} workflows={searchResults()}
                                                      fetchWorkflows={fetchWorkflows}/>
                                        :
                                        <div className="workflows-list-item-no">
                                            <div style={{alignSelf: "center"}}>
                                                <div style={{padding: "5px"}}>
                                                    <span>No workflows are saved.</span>
                                                </div>
                                            </div>

                                        </div>
                                    }
                                </>
                            }
                        </>}

                </Col>
                {searchResults().length > 0 ?
                    <Col style={{marginTop: "30px"}}>
                        <ReactPaginate
                            breakClassName="page-item"
                            breakLinkClassName="page-link"
                            previousClassName="page-item"
                            previousLinkClassName="page-link"
                            nextClassName="page-item"
                            nextLinkClassName="page-link"
                            previousLabel="<"
                            containerClassName="pagination"
                            pageClassName="page-item"
                            pageLinkClassName="page-link"
                            nextLabel=">"
                            onPageChange={(num) => {
                                setPagination({...pagination, offset: num.selected * 10})
                            }}
                            pageCount={Math.ceil(pagination.total / 10)}
                        />
                    </Col> : ""}
            </Row>
        </>
    );
}