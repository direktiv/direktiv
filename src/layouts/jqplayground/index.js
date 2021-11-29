import { useJQPlayground } from 'direktiv-react-hooks';
import { useEffect, useState } from 'react';
import { IoLink, IoReload } from 'react-icons/io5';
import { VscFileCode } from 'react-icons/vsc';
import Button from '../../components/button';
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import DirektivEditor from '../../components/editor';
import FlexBox from '../../components/flexbox';
import HelpIcon from '../../components/help';
import { Config } from '../../util';
import './style.css';


export default function JQPlayground() {

    const [filter, setFilter] = useState(".")
    const [input, setInput] = useState(JSON.stringify({}, null, 2))

    const {data, err, executeJQ, cheatSheet} = useJQPlayground(Config.url)

    if(err){
        // jq query went busted
        console.log(err, "jq err")
    }

    return(
        <FlexBox id="jq-page" className="col gap" style={{paddingRight:"8px"}}>
            <JQFilter data={input} query={filter} setFilter={setFilter} executeJQ={executeJQ}/>
            <FlexBox className="gap col" >
                <FlexBox className="gap wrap">
                    <FlexBox style={{minWidth:"380px"}}>
                        <JQInput input={input} setInput={setInput}/>
                    </FlexBox>
                    <FlexBox style={{minWidth:"380px"}}>
                        <JQOutput data={data}/>                    
                    </FlexBox>
                </FlexBox>
            </FlexBox>
            <FlexBox className="gap col" >
                <FlexBox className="gap wrap">
                    <HowToJQ />
                    <ExamplesJQ cheatSheet={cheatSheet} setFilter={setFilter} setInput={setInput} executeJQ={executeJQ}/>
                    {/* <HowToJQ />
                    <ExamplesJQ cheatSheet={cheatSheet} setFilter={setFilter} setInput={setInput} executeJQ={executeJQ}/> */}
                </FlexBox>
            </FlexBox>
        </FlexBox>
    )
}

function HowToJQ(){
    return(
        <FlexBox style={{ maxWidth: "380px"}}>
            <ContentPanel style={{width:"100%"}}>
                <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <VscFileCode/>
                        </ContentPanelTitleIcon>
                        <FlexBox className="gap" style={{ alignItems: "center"}}>
                            <div>
                                How it works
                            </div>
                            <HelpIcon msg={"Brief instructions on how JQ Playground works"} />
                        </FlexBox>
                    </ContentPanelTitle>
                    <ContentPanelBody>
                        <FlexBox className="col gap" style={{fontSize:"10pt"}}>
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
                            <Button className="reveal-btn small shadow">
                                                    <FlexBox className="gap">
                                                        <IoLink className="auto-margin" />
                                                        <a href="https://stedolan.github.io/jq/manual/">
                                                            View JQ Manual
                                                        </a>
                                                    </FlexBox>
                                                </Button>
                        </FlexBox>
                    </ContentPanelBody>
            </ContentPanel>
        </FlexBox>
    )
}

function ExamplesJQ(props){
    const {cheatSheet, setFilter, setInput, executeJQ} = props

    async function loadJQ(f, i) {
        setFilter(f)
        setInput(JSON.stringify(JSON.parse(i), null, 2))
        await executeJQ(f, btoa(i))
    }

    const half = Math.ceil(cheatSheet.length / 2);    

    const firstHalf = cheatSheet.slice(0, half)
    const secondHalf = cheatSheet.slice(-half)

    return(
        <FlexBox style={{flex: 2}}>
            <ContentPanel style={{minHeight:"280px"}}>
                <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <VscFileCode/>
                        </ContentPanelTitleIcon>
                        <FlexBox className="gap" style={{ alignItems: "center" }}>
                            <div>
                                Cheatsheet
                            </div>
                            <HelpIcon msg={"A list of examples that you can load into the playground"} />
                        </FlexBox>
                    </ContentPanelTitle>
                    <ContentPanelBody >

                        <table >
                            <tbody>
                                {firstHalf.map((obj)=>{
                                    console.log(obj)
                                    return(
                                        <tr>
                                            <td className="jq-example">
                                                {obj.example}
                                            </td>
                                            <td>
                                                {obj.tip}
                                            </td>
                                            <td onClick={()=>loadJQ(obj.filter, obj.json)}>
                                                <Button className="reveal-btn small shadow">
                                                    <FlexBox className="gap">
                                                        <IoReload className="auto-margin" />
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
                        <table>
         <tbody>
                                {secondHalf.map((obj)=>{
                                    console.log(obj)
                                    return(
                                        <tr>
                                            <td className="jq-example">
                                                {obj.example}
                                            </td>
                                            <td>
                                                {obj.tip}
                                            </td>
                                            <td onClick={()=>loadJQ(obj.filter, obj.json)}>
                                                <Button className="reveal-btn small shadow">
                                                    <FlexBox className="gap">
                                                        <IoReload className="auto-margin" />
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
                    <FlexBox className="gap" style={{ alignItems: "center" }}>
                        <div>
                            Output
                        </div>
                        <HelpIcon msg={"The output of the JQ query"} />
                    </FlexBox>
                </ContentPanelTitle>
                <ContentPanelBody >
                    <FlexBox style={{overflow:"hidden" , height: "422px", maxHeight:"422px"}}>
                        <DirektivEditor value={output} height="100%" dlang={"json"} />
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
                    <FlexBox className="gap" style={{ alignItems: "center" }}>
                        <div>
                            Input
                        </div>
                        <HelpIcon msg={"The input to feed the JQ query"} />
                    </FlexBox>
                </ContentPanelTitle>
                <ContentPanelBody >
                    <FlexBox style={{overflow:"hidden" , height: "422px", maxHeight:"422px"}}>
                        <DirektivEditor value={input} setDValue={setInput}  height="100%" dlang={"json"}/>
                    </FlexBox>
                </ContentPanelBody>
        </ContentPanel>
    )
}

function JQFilter(props) {
    const {data, setFilter, executeJQ, query} = props

    return(
        <FlexBox style={{ maxHeight:"105px"}}>
            <ContentPanel style={{width:"100%"}}>
                <ContentPanelTitle>
                    <ContentPanelTitleIcon>
                        <VscFileCode/>
                    </ContentPanelTitleIcon>
                    <FlexBox className="gap" style={{ alignItems: "center" }}>
                        <div>
                            JQ Filter
                        </div>
                        <HelpIcon msg={"A simple JQ playground to test your queries"} />
                    </FlexBox>
                </ContentPanelTitle>
                <ContentPanelBody >
                    <FlexBox className="gap wrap" style={{height:"40px"}}>
                        <FlexBox style={{fontSize: "12pt"}} >
                            <input style={{height:"28px", width:"100%"}} onChange={(e)=>setFilter(e.target.value)} value={query} placeholder={"Enter a Filter to JQ on"} type="text" />
                        </FlexBox>
                        <FlexBox style={{maxWidth:"65px"}}>
                             
                            <Button className="small" onClick={()=>executeJQ(query, btoa(data))}>
                                Execute
                            </Button>
                        </FlexBox>
                    </FlexBox>
                </ContentPanelBody>
            </ContentPanel>
        </FlexBox>
    )
}