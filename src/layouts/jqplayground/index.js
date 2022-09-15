import { useJQPlayground } from 'direktiv-react-hooks';
import { useCallback, useEffect, useState } from 'react';
import { VscFileCode, VscArrowRight } from 'react-icons/vsc';
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import DirektivEditor from '../../components/editor';
import Alert from '../../components/alert';
import FlexBox from '../../components/flexbox';
import HelpIcon from '../../components/help';
import { Config } from '../../util';
import './style.css';
import Button from '../../components/button';


export default function JQPlayground() {

    const [filter, setFilter] = useState(localStorage.getItem('jqFilter') ? localStorage.getItem('jqFilter') : ".")
    const [input, setInput] = useState(localStorage.getItem('jqInput') ? localStorage.getItem('jqInput') : JSON.stringify({}, null, 2))
    const [error, setError] = useState(null)



    const {data, err, executeJQ, cheatSheet} = useJQPlayground(Config.url, localStorage.getItem("apikey"))

    const executeAndSave = useCallback((...args)=>{
        localStorage.setItem('jqInput', input)
        localStorage.setItem('jqFilter', filter)
        return executeJQ(...args)
    }, [executeJQ, filter, input])

    // Save state every 2 seconds
    useEffect(()=>{
        if (filter == null || input == null ) {
            return
        }
        
        let timer = setInterval(async ()=>{
            localStorage.setItem('jqInput', input)
            localStorage.setItem('jqFilter', filter)
        }, 2000)

        return function cleanup() {
            clearInterval(timer)
        }
    },[filter,input])


    useEffect(() => {
        setError(err)
    }, [err])
 
    return(
        <FlexBox id="jq-page" className="col gap" style={{paddingRight:"8px"}}>
            <JQFilter data={input} query={filter} error={error} setFilter={setFilter} executeJQ={executeAndSave} setError={setError}/>
            <FlexBox col gap >
                <FlexBox gap wrap>
                    <FlexBox style={{minWidth:"380px"}}>
                        <JQInput input={input} setInput={setInput}/>
                    </FlexBox>
                    <FlexBox style={{minWidth:"380px"}}>
                        <JQOutput data={data}/>                    
                    </FlexBox>
                </FlexBox>
            </FlexBox>
            <FlexBox col gap >
                <FlexBox className="gap box-wrap">
                    <HowToJQ />
                    <ExamplesJQ cheatSheet={cheatSheet} setFilter={setFilter} setInput={setInput} executeJQ={executeAndSave} setError={setError}/>
                </FlexBox>
            </FlexBox>
        </FlexBox>
    )
}

function HowToJQ(){
    return(
        <FlexBox className="how-to-jq">
            <ContentPanel style={{width:"100%"}}>
                <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <VscFileCode/>
                        </ContentPanelTitleIcon>
                        <FlexBox gap style={{ alignItems: "center"}}>
                            <div>
                                How it works
                            </div>
                            <HelpIcon msg={"Brief instructions on how JQ Playground works"} />
                        </FlexBox>
                    </ContentPanelTitle>
                    <ContentPanelBody>
                        <FlexBox col gap style={{fontSize:"10pt"}}>
                            <span style={{fontWeight:"bold"}}>JQ Playground is an envrioment where you can quickly test your jq commands against JSON.</span>
                            <span>There are two inputs in the playground:</span>
                            <ul>
                                <li><span style={{fontWeight:"bold"}}>Filter</span> - This is the jq command that will be used to transform your JSON input</li>
                                <li><span style={{fontWeight:"bold"}}>JSON</span> - This is the JSON input that will be transformed</li>
                            </ul>
                            <div>
                                The transformed JSON is shown in the Result output field.
                            </div>
                            <div>
                                For information on the JQ syntax, please refer to the offical JQ manual online.
                            </div>
                            <div>
                                <Button variant='outlined' color="info">
                                    <FlexBox gap>
                                        <VscArrowRight className="auto-margin" />
                                        <a href="https://stedolan.github.io/jq/manual/">
                                            View JQ Manual
                                        </a>
                                    </FlexBox>
                                </Button>
                            </div>
                        </FlexBox>
                    </ContentPanelBody>
            </ContentPanel>
        </FlexBox>
    )
}

function ExamplesJQ(props){
    const {cheatSheet, setFilter, setInput, executeJQ, setError} = props

    async function loadJQ(f, i) {
        setError(null)
        setFilter(f)
        setInput(JSON.stringify(JSON.parse(i), null, 2))
        await executeJQ(f, btoa(i))
    }

    const half = Math.ceil(cheatSheet.length / 2);    

    const firstHalf = cheatSheet.slice(0, half)
    const secondHalf = cheatSheet.slice(-half)

    return(
        <FlexBox style={{flex: 2}}>
            <ContentPanel style={{minHeight:"280px", width: "100%"}}>
                <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <VscFileCode/>
                        </ContentPanelTitleIcon>
                        <FlexBox gap style={{ alignItems: "center" }}>
                            <div>
                                Cheatsheet
                            </div>
                            <HelpIcon msg={"A list of examples that you can load into the playground"} />
                        </FlexBox>
                    </ContentPanelTitle>
                    <ContentPanelBody >

                        <table style={{ width: "50%", fontSize:"10pt"}}>
                            <tbody>
                                {firstHalf.map((obj)=>{
                                    return(
                                        <tr>
                                            <td className="jq-example" style={{ width: "25%"}}>
                                                {obj.example}
                                            </td>
                                            <td>
                                                {obj.tip}
                                            </td>
                                            <td style={{ width: "20%"}} onClick={()=>loadJQ(obj.filter, obj.json)}>
                                                <Button variant='outlined' color="info">
                                                    <FlexBox gap>
                                                        <VscFileCode className="auto-margin" />
                                                        <div>
                                                            Load
                                                        </div>
                                                    </FlexBox>
                                                </Button>
                                            </td>
                                        </tr>
                                    )
                                })}
                            </tbody>
                            
                        </table>
                        <table style={{ width: "50%", fontSize:"10pt"}}>
         <tbody>
                                {secondHalf.map((obj)=>{
                                    return(
                                        <tr>
                                            <td style={{ width: "25%"}} className="jq-example">
                                                {obj.example}
                                            </td>
                                            <td>
                                                {obj.tip}
                                            </td>
                                            <td style={{ width: "20%"}} onClick={()=>loadJQ(obj.filter, obj.json)}>
                                                <Button variant='outlined' color="info">
                                                    <FlexBox gap>
                                                        <VscFileCode className="auto-margin" />
                                                        <div>
                                                            Load
                                                        </div>
                                                    </FlexBox>
                                                </Button>
                                            </td>
                                        </tr>
                                    )
                                })}
                            </tbody>
                        </table>
                    </ContentPanelBody>
            </ContentPanel>
        </FlexBox>
    )
}

function JQOutput(props) {
    const {data} = props

    const [output, setOutput] = useState("")

    useEffect(()=>{
        if(data !== output){
            if (data){
                setOutput(data.toString())
            }
        }
    }, [data, output])

    return(
        <ContentPanel style={{width:"100%"}}>
            <ContentPanelTitle>
                    <ContentPanelTitleIcon>
                        <VscFileCode/>
                    </ContentPanelTitleIcon>
                    <FlexBox gap style={{ alignItems: "center" }}>
                        <div>
                            Output
                        </div>
                        <HelpIcon msg={"The output of the JQ query"} />
                    </FlexBox>
                </ContentPanelTitle>
                <ContentPanelBody >
                    <FlexBox style={{overflow:"hidden" , height: "422px", maxHeight:"422px"}}>
                        <DirektivEditor readonly={true} value={output} dlang={"json"} />
                    </FlexBox>
                </ContentPanelBody>
        </ContentPanel>
    )
}

function JQInput(props) {
    const {input, setInput} = props
    return(
        <ContentPanel style={{width:"100%"}}>
            <ContentPanelTitle>
                    <ContentPanelTitleIcon>
                        <VscFileCode/>
                    </ContentPanelTitleIcon>
                    <FlexBox gap style={{ alignItems: "center" }}>
                        <div>
                            Input
                        </div>
                        <HelpIcon msg={"The input to feed the JQ query"} />
                    </FlexBox>
                </ContentPanelTitle>
                <ContentPanelBody >
                    <FlexBox style={{overflow:"hidden" , height: "422px", maxHeight:"422px"}}>
                        <DirektivEditor readonly={false} value={input} setDValue={setInput}  dlang={"json"}/>
                    </FlexBox>
                </ContentPanelBody>
        </ContentPanel>
    )
}

function JQFilter(props) {
    const {data, setFilter, setError, executeJQ, query, error} = props
    
    async function execute() {
        setError(null)
        // setFilter("")
        try{
            const result = await executeJQ(query, btoa(data))
            return {
                error: false,
                data: result
            }
        }catch(e){
            setError(e.toString())
            return {
                error: true,
                msg: e.toString()
            }
        }
        
    }


    return(
        <FlexBox style={{ maxHeight:"205px" }}>
            <ContentPanel style={{width:"100%"}}>
                <ContentPanelTitle>
                    <ContentPanelTitleIcon>
                        <VscFileCode/>
                    </ContentPanelTitleIcon>
                    <FlexBox gap style={{ alignItems: "center" }}>
                        <div>
                            JQ Filter
                        </div>
                        <HelpIcon msg={"A simple JQ playground to test your queries"} />
                    </FlexBox>
                </ContentPanelTitle>
                <ContentPanelBody >
                    <FlexBox className="gap wrap center-y" style={{height:"40px"}}>
                        <FlexBox style={{fontSize: "12pt"}} >
                            <input style={{height:"28px", width:"100%"}} onChange={(e)=>setFilter(e.target.value)} value={query} placeholder={"Enter a Filter to JQ on"} type="text" />
                        </FlexBox>
                        <FlexBox style={{maxWidth:"65px"}}>
                            <Button onClick={()=>execute()}>
                                Execute
                            </Button>
                        </FlexBox>
                    </FlexBox>
                </ContentPanelBody>
                <FlexBox>
                {error ? <Alert severity="error" variant="filled" grow onClose={()=>{setError(null)}}><div><span>error executing JQ command:</span>{error.replace("execute jq: error executing JQ command:", "")}</div></Alert> : null}
                </FlexBox>
            </ContentPanel>
        </FlexBox>
    )
}