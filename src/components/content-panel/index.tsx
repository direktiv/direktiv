import React from 'react'
import FlexBox from '../flexbox';
import './style.css';

export interface ContentPanelProps extends React.HTMLAttributes<HTMLDivElement> {
    /**
    * If true, expand to available space when inside a flex container by setting flex-grow to 1 
    */
    grow?: boolean
}

/**
* Primary "card" designed container used in direktiv UI. This component "ContentPanel" has four children jsx items used to structure its contents.
* * ContentPanelTitle - Header bar of container used for title.
* * ContentPanelTitleIcon - Child Item of ContentPanelTitle and is used for icons in the Header.
* * ContentPanelBody - Main body of content.
* * ContentPanelFooter - Footer bar of container.
*/
function ContentPanel({
    grow = false,
    ...props
}: ContentPanelProps) {
    return (
        <div 
        {...props} style={{ display: "flex", flexDirection: "column", flexGrow: grow ? "1" : undefined, ...props.style}} className={`content-panel-parent opaque ${props.className ? props.className : ""}`}/>
    );
}

export default ContentPanel;

export function ContentPanelTitle({ ...props }: React.HTMLAttributes<HTMLDivElement>) {
    return (
        <FlexBox {...props} className={`content-panel-title ${props.className ? props.className : ""}`} />
    );
}

export function ContentPanelTitleIcon({ ...props }: React.HTMLAttributes<HTMLDivElement>) {
    return (
        <div {...props} className={`content-panel-title-icon ${props.className ? props.className : ""}`} />
    );
}

export function ContentPanelBody({ ...props }: React.HTMLAttributes<HTMLDivElement>) {
    return (
        <div {...props} className={`content-panel-body ${props.className ? props.className : ""}`} />
    );
}

export function ContentPanelFooter({ ...props }: React.HTMLAttributes<HTMLDivElement>) {
    return (
        <div {...props} className={`content-panel-footer ${props.className ? props.className : ""}`} />
    );
}