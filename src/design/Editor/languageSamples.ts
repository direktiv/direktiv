export const HtmlSample = `
<!DOCTYPE HTML>
<!--Example of comments in HTML-->
<html>
<head>
	<!--This is the head section-->
	<title>HTML Sample</title>
	<meta charset="utf-8">

	<!--This is the style tag to set style on elements-->
	<style type="text/css">
		h1
		{
			font-family: Tahoma;
			font-size: 40px;
			font-weight: normal;
			margin: 50px;
			color: #a0a0a0;
		}

		h2
		{
			font-family: Tahoma;
			font-size: 30px;
			font-weight: normal;
			margin: 50px;
			color: #fff;
		}

		p
		{
			font-family: Tahoma;
			font-size: 17px;
			font-weight: normal;
			margin: 0px 200px;
			color: #fff;
		}

		div.Center
		{
			text-align: center;
		}

		div.Blue
		{
			padding: 50px;
			background-color: #7bd2ff;
		}

		button.Gray
		{
			font-family: Tahoma;
			font-size: 17px;
			font-weight: normal;
			margin-top: 100px;
			padding: 10px 50px;
			background-color: #727272;
			color: #fff;
			outline: 0;
    			border: none;
    			cursor: pointer;
		}

		button.Gray:hover
		{
			background-color: #898888;
		}

		button.Gray:active
		{
			background-color: #636161;
		}

	</style>

	<!--This is the script tag-->
	<script type="text/javascript">
		function ButtonClick(){
			// Example of comments in JavaScript
			window.alert("I'm an alert sample!");
		}
	</script>
</head>
<body>
	<!--This is the body section-->
	<div class="Center">
		<h1>NAME OF SITE</h1>
	</div>
	<div class="Center Blue">
			<h2>I'm h2 Header! Edit me in &lt;h2&gt;</h2>
			<p>
				I'm a paragraph! Edit me in &lt;p&gt;
				to add your own content and make changes to the style and font.
				It's easy! Just change the text between &lt;p&gt; ... &lt;/p&gt; and change the style in &lt;style&gt;.
				You can make it as long as you wish. The browser will automatically wrap the lines to accommodate the
				size of the browser window.
			</p>
			<button class="Gray" onclick="ButtonClick()">Click Me!</button>
	</div>
</body>
</html>
`;
export const CSSSample = `
html {
    background-color: #e2e2e2;
    margin: 0;
    padding: 0;
}

body {
    background-color: #fff;
    border-top: solid 10px #000;
    color: #333;
    font-size: .85em;
    font-family: "Segoe UI","HelveticaNeue-Light", sans-serif;
    margin: 0;
    padding: 0;
}

a:link, a:visited,
a:active, a:hover {
    color: #333;
    outline: none;
    padding-left: 0;
    padding-right: 3px;
    text-decoration: none;

}


a:hover {
    background-color: #c7d1d6;
}


header, footer, hgroup
nav, section {
    display: block;
}

.float-left {
    float: left;
}

.float-right {
    float: right;
}

.highlight {
/*    background-color: #a6dbed;
    padding-left: 5px;
    padding-right: 5px;*/
}

.clear-fix:after {
    content: ".";
    clear: both;
    display: block;
    height: 0;
    visibility: hidden;
}

h1, h2, h3,
h4, h5, h6 {
    color: #000;
    margin-bottom: 0;
    padding-bottom: 0;

}

h1 {
    font-size: 2em;
}

h2 {
    font-size: 1.75em;
}

h3 {
    font-size: 1.2em;
}

h4 {
    font-size: 1.1em;
}

h5, h6 {
    font-size: 1em;
}


.tile {
    /* 2px solid #7ac0da; */
    border: 0;

    float: left;
    width: 200px;
    height: 325px;

    padding: 5px;
    margin-right: 5px;
    margin-bottom: 20px;
    margin-top: 20px;
    -webkit-perspective: 0;
    -webkit-transform-style: preserve-3d;
    -webkit-transition: -webkit-transform 0.2s;
    -webkit-box-shadow: 0 1px 1px rgba(0,0,0,0.3);
    background-position: center center;
    background-repeat: no-repeat;

    background-color:  #fff;
}

.tile-item {
    /* 2px solid #7ac0da; */
    border-color: inherit;
    float: left;
    width: 50px;
    height: 70px;
    margin-right: 20px;
    margin-bottom: 20px;
    margin-top: 20px;
    background-image: url('../Images/documents.png');
    background-repeat: no-repeat;

}

.tile-wrapper {
    width: 100%;
    font-family: "Segoe UI" , Tahoma, Geneva, Verdana, sans-serif;
    line-height: 21px;
    font-size: 14px;
}

a.blue-box {
    font-size: 28px;
    height: 100px;
    display: block;
    border-style: solid;
    border-width: 1px 1px 4px 1px;
    border-color: #C0C0C0 #C0C0C0 #8ABAE4 #C0C0C0;
    padding-top: 15px;
    padding-left: 15px;
}

    a.blue-box:hover {
    border: 4px solid #8ABAE4;
    padding-top: 12px;
    padding-left: 12px;
    background-color: #FFFFFF;
}

a.green-box {
    font-size: 28px;
    height: 100px;
    display: block;
    border-style: solid;
    border-width: 1px 1px 4px 1px;
    border-color: #C0C0C0 #C0C0C0 #9CCF42 #C0C0C0;
    padding-top: 15px;
    padding-left: 15px;
}

    a.green-box:hover {
        border: 4px solid #9CCF42;
        padding-top: 12px;
        padding-left: 12px;
        background-color: #FFFFFF;
}


a.green-box2 {
    font-size: 14px;
    height: 48px;
    width: 48px;
    display: block; /* border-color: #C0C0C0; */
    padding-top: 6px;
    font-weight: bold;

}

    a.green-box2:hover {
    border: solid #8ABAE4;
    padding-top: 0px;
    padding-left: 0px;
    background-image: url('../Images/documents.png');
    background-color: #EFEFEF;
}

a.yellow-box {
    font-size: 28px;
    height: 100px;
    display: block;
    border-style: solid;
    border-width: 1px 1px 4px 1px;
    border-color: #C0C0C0 #C0C0C0 #DECF6B #C0C0C0;
    padding-top: 15px;
    padding-left: 15px;
}

    a.yellow-box:hover {
        border: 4px solid #DECF6B;
        padding-top: 12px;
        padding-left: 12px;
        background-color: #FFFFFF;
}


a.red-box {
    font-size: 28px;
    height: 100px;
    display: block;
    border-style: solid;
    border-width: 1px 1px 4px 1px;
    border-color: #C0C0C0 #C0C0C0 #F79E84 #C0C0C0;
    padding-top: 15px;
    padding-left: 15px;
}

    a.red-box:hover {
    border: 4px solid #F79E84;
    padding-top: 12px;
    padding-left: 12px;
    background-color: #FFFFFF;
}

/* main layout
----------------------------------------------------------*/
.content-wrapper {
    margin: 0 auto;
    max-width: 960px;
}

#body {
    background-color: #efeeef;
    clear: both;
    padding-bottom: 35px;
}

    .main-content {
        background: url("../images/accent.png") no-repeat;
        padding-left: 10px;
        padding-top: 30px;
    }

    .featured + .main-content {
        background: url("../images/heroaccent.png") no-repeat;
    }

footer {
    clear: both;
    background-color: #e2e2e2;
    font-size: .8em;
    height: 100px;
}


/* site title
----------------------------------------------------------*/
.site-title {
    color: #0066CC; /* font-family: Rockwell, Consolas, "Courier New", Courier, monospace; */
    font-size: 3.3em;
    margin-top: 40px;
    margin-bottom: 0;
}

.site-title a, .site-title a:hover, .site-title a:active  {
    background: none;
    color: #0066CC;
    outline: none;
    text-decoration: none;
}


/* login
----------------------------------------------------------*/
#login {
    display: block;
    font-size: .85em;
    margin-top: 20px;
    text-align: right;
}

    #login a {
        background-color: #d3dce0;
        margin-left: 10px;
        margin-right: 3px;
        padding: 2px 3px;
        text-decoration: none;
    }

    #login a.username {
        background: none;
        margin-left: 0px;
        text-decoration: underline;
    }

    #login li {
        display: inline;
        list-style: none;
    }


/* menu
----------------------------------------------------------*/
ul#menu {
    font-size: 1.3em;
    font-weight: 600;
    margin: 0;
    text-align: right;
            text-decoration: none;

}

    ul#menu li {
        display: inline;
        list-style: none;
        padding-left: 15px;
    }

        ul#menu li a {
            background: none;
            color: #999;
            text-decoration: none;
        }

        ul#menu li a:hover {
            color: #333;
            text-decoration: none;
        }



/* page elements
----------------------------------------------------------*/
/* featured */
.featured {
    background-color: #fff;
}

    .featured .content-wrapper {
        /*background-color: #7ac0da;
        background-image: -ms-linear-gradient(left, #7AC0DA 0%, #A4D4E6 100%);
        background-image: -o-linear-gradient(left, #7AC0DA 0%, #A4D4E6 100%);
        background-image: -webkit-gradient(linear, left top, right top, color-stop(0, #7AC0DA), color-stop(1, #A4D4E6));
        background-image: -webkit-linear-gradient(left, #7AC0DA 0%, #A4D4E6 100%);
        background-image: linear-gradient(left, #7AC0DA 0%, #A4D4E6 100%);
        color: #3e5667;
        */
        padding:  0px 40px 30px 40px;
    }

        .featured hgroup.title h1, .featured hgroup.title h2 {
            /* color: #fff;
                */
        }

        .featured p {
            font-size: 1.1em;
        }

/* page titles */
hgroup.title {
    margin-bottom: 10px;
}

hgroup.title h1, hgroup.title h2 {
display: inline;
}

hgroup.title h2 {
    font-weight: normal;
}

/* releases */
.milestone {
    color: #fff;
    background-color: #8ABAE4;
    font-weight:  normal;
    padding:  10px 10px 10px 10px;
    margin: 0 0 0 0;
}
    .milestone .primary {
        font-size: 1.75em;
    }

    .milestone .secondary {
    font-size: 1.2em;
    font-weight: normal;
    /* padding: 5px 5px 5px 10px;*/
        }

/* features */
section.feature {
    width: 200px;
    float: left;
    padding: 10px;
}

/* ordered list */
ol.round {
    list-style-type: none;
    padding-left: 0;
}

    ol.round li {
        margin: 25px 0;
        padding-left: 45px;
    }

        ol.round li.one {
            background: url("../images/orderedlistOne.png") no-repeat;
        }

        ol.round li.two {
            background: url("../images/orderedlistTwo.png") no-repeat;
        }

        ol.round li.three {
            background: url("../images/orderedlistThree.png") no-repeat;
        }

/* content */
article {
    float: left;
    width: 70%;
}

aside {
    float: right;
    width: 25%;
}

    aside ul {
        list-style: none;
        padding: 0;
    }

     aside ul li {
        background: url("../images/bullet.png") no-repeat 0 50%;
        padding: 2px 0 2px 20px;
     }

.label {
    font-weight: 700;
}

/* login page */
#loginForm {
    border-right: solid 2px #c8c8c8;
    float: left;
    width: 45%;
}

    #loginForm .validation-error {
        display: block;
        margin-left: 15px;
    }

#socialLoginForm {
    margin-left: 40px;
    float: left;
    width: 50%;
}

/* contact */
.contact h3 {
    font-size: 1.2em;
}

.contact p {
    margin: 5px 0 0 10px;
}

.contact iframe {
    border: solid 1px #333;
    margin: 5px 0 0 10px;
}

/* forms */
fieldset {
    border: none;
    margin: 0;
    padding: 0;
}

    fieldset legend {
        display: none;
    }

    fieldset ol {
        padding: 0;
        list-style: none;
    }

        fieldset ol li {
            padding-bottom: 5px;
        }

    fieldset label {
        display: block;
        font-size: 1.2em;
        font-weight: 600;
    }

    fieldset label.checkbox {
        display: inline;
    }

    fieldset input[type="text"],
    fieldset input[type="password"] {
        border: 1px solid #e2e2e2;
        color: #333;
        font-size: 1.2em;
        margin: 5px 0 6px 0;
        padding: 5px;
        width: 300px;
    }

        fieldset input[type="text"]:focus,
        fieldset input[type="password"]:focus {
            border: 1px solid #7ac0da;
        }

    fieldset input[type="submit"] {
        background-color: #d3dce0;
        border: solid 1px #787878;
        cursor: pointer;
        font-size: 1.2em;
        font-weight: 600;
        padding: 7px;
    }

/* ajax login/registration dialog */
.modal-popup {
    font-size: 0.7em;
}

/* info and errors */
.message-info {
    border: solid 1px;
    clear: both;
    padding: 10px 20px;
}

.message-error {
    clear: both;
    color: #e80c4d;
    font-size: 1.1em;
    font-weight: bold;
    margin: 20px 0 10px 0;
}

.message-success {
    color: #7ac0da;
    font-size: 1.3em;
    font-weight: bold;
    margin: 20px 0 10px 0;
}

.success {
    color: #7ac0da;
}

.error {
    color: #e80c4d;
}

/* styles for validation helpers */
.field-validation-error {
    color: #e80c4d;
    font-weight: bold;
}

.field-validation-valid {
    display: none;
}

input[type="text"].input-validation-error,
input[type="password"].input-validation-error {
    border: solid 1px #e80c4d;
}

.validation-summary-errors {
    color: #e80c4d;
    font-weight: bold;
    font-size: 1.1em;
}

.validation-summary-valid {
    display: none;
}


/* social */
ul#social li {
    display: inline;
    list-style: none;
}

    ul#social li a {
        color: #999;
        text-decoration: none;
    }

    a.facebook, a.twitter {
        display: block;
        float: left;
        height: 24px;
        padding-left: 17px;
        text-indent: -9999px;
        width: 16px;
    }

    a.facebook {
        background: url("../images/facebook.png") no-repeat;
    }

    a.twitter {
        background: url("../images/twitter.png") no-repeat;
    }



/********************
*   Mobile Styles   *
********************/
@media only screen and (max-width: 850px) {

    /* header
    ----------------------------------------------------------*/
    header .float-left,
    header .float-right {
        float: none;
    }

    /* logo */
    header .site-title {
        /*margin: 0; */
        /*margin: 10px;*/
        text-align: left;
        padding-left: 0;
    }

    /* login */
    #login {
        font-size: .85em;
        margin-top: 0;
        text-align: center;
    }

        #login ul {
            margin: 5px 0;
            padding: 0;
        }

        #login li {
            display: inline;
            list-style: none;
            margin: 0;
            padding:0;
        }

        #login a {
            background: none;
            color: #999;
            font-weight: 600;
            margin: 2px;
            padding: 0;
        }

        #login a:hover {
            color: #333;
        }

    /* menu */
    nav {
        margin-bottom: 5px;
    }

    ul#menu {
        margin: 0;
        padding:0;
        text-align: center;
    }

        ul#menu li {
            margin: 0;
            padding: 0;
        }


    /* main layout
    ----------------------------------------------------------*/
    .main-content,
    .featured + .main-content {
        background-position: 10px 0;
    }

    .content-wrapper {
        padding-right: 10px;
        padding-left: 10px;
    }

    .featured .content-wrapper {
        padding: 10px;
    }

    /* page content */
    article, aside {
        float: none;
        width: 100%;
    }

    /* ordered list */
    ol.round {
        list-style-type: none;
        padding-left: 0;
    }

        ol.round li {
            padding-left: 10px;
            margin: 25px 0;
        }

            ol.round li.one,
            ol.round li.two,
            ol.round li.three {
                background: none;
            }

     /* features */
     section.feature {
        float: none;
        padding: 10px;
        width: auto;
     }

        section.feature img {
            color: #999;
            content: attr(alt);
            font-size: 1.5em;
            font-weight: 600;
        }

    /* forms */
    fieldset input[type="text"],
    fieldset input[type="password"] {
        width: 90%;
    }

    /* login page */
    #loginForm {
        border-right: none;
        float: none;
        width: auto;
    }

        #loginForm .validation-error {
            display: block;
            margin-left: 15px;
        }

    #socialLoginForm {
        margin-left: 0;
        float: none;
        width: auto;
    }

    /* footer
    ----------------------------------------------------------*/
    footer .float-left,
    footer .float-right {
        float: none;
    }

    footer {
        text-align: center;
        height: auto;
        padding: 10px 0;
    }

        footer p {
            margin: 0;
        }

        ul#social {
            padding:0;
            margin: 0;
        }

         a.facebook, a.twitter {
            background: none;
            display: inline;
            float: none;
            height: auto;
            padding-left: 0;
            text-indent: 0;
            width: auto;
        }
}

.subsite {
	color: #444;
}

h3 {
	font-weight: normal;
	font-size: 24px;
	color: #444;
	margin-bottom: 20px;
}

.tiles {
	padding-bottom: 20px;
	background-color: #e3e3e3;
}

#editor {
	margin: 0 auto;
	height: 500px;
	border: 1px solid #ccc;
}

.monaco-editor.monaco, .monaco-editor.vs, .monaco-editor.eclipse {
	background: #F9F9F9;
}

.monaco-editor.monaco .monaco-editor-background, .monaco-editor.vs .monaco-editor-background, .monaco-editor.eclipse .monaco-editor-background {
	background: #F9F9F9;
}
`;
export const JsonSample = `
{
	"type": "team",
	"test": {
		"testPage": "tools/testing/run-tests.htm",
		"enabled": true
	},
    "search": {
        "excludeFolders": [
			".git",
			"node_modules",
			"tools/bin",
			"tools/counts",
			"tools/policheck",
			"tools/tfs_build_extensions",
			"tools/testing/jscoverage",
			"tools/testing/qunit",
			"tools/testing/chutzpah",
			"server.net"
        ]
    },
	"languages": {
		"vs.languages.typescript": {
			"validationSettings": [{
				"scope":"/",
				"noImplicitAny":true,
				"noLib":false,
				"extraLibs":[],
				"semanticValidation":true,
				"syntaxValidation":true,
				"codeGenTarget":"ES5",
				"moduleGenTarget":"",
				"lint": {
                    "emptyBlocksWithoutComment": "warning",
                    "curlyBracketsMustNotBeOmitted": "warning",
                    "comparisonOperatorsNotStrict": "warning",
                    "missingSemicolon": "warning",
                    "unknownTypeOfResults": "warning",
                    "semicolonsInsteadOfBlocks": "warning",
                    "functionsInsideLoops": "warning",
                    "functionsWithoutReturnType": "warning",
                    "tripleSlashReferenceAlike": "warning",
                    "unusedImports": "warning",
                    "unusedVariables": "warning",
                    "unusedFunctions": "warning",
                    "unusedMembers": "warning"
                }
			},
			{
				"scope":"/client",
				"baseUrl":"/client",
				"moduleGenTarget":"amd"
			},
			{
				"scope":"/server",
				"moduleGenTarget":"commonjs"
			},
			{
				"scope":"/build",
				"moduleGenTarget":"commonjs"
			},
			{
				"scope":"/node_modules/nake",
				"moduleGenTarget":"commonjs"
			}],
			"allowMultipleWorkers": true
		}
	}
}
`;
export const ShellSample = `
#!/bin/bash
# Simple line count example, using bash
#
# Bash tutorial: http://linuxconfig.org/Bash_scripting_Tutorial#8-2-read-file-into-bash-array
# My scripting link: http://www.macs.hw.ac.uk/~hwloidl/docs/index.html#scripting
#
# Usage: ./line_count.sh file
# -----------------------------------------------------------------------------

# Link filedescriptor 10 with stdin
exec 10<&0
# stdin replaced with a file supplied as a first argument
exec < $1
# remember the name of the input file
in=$1

# init
file="current_line.txt"
let count=0

# this while loop iterates over all lines of the file
while read LINE
do
    # increase line counter
    ((count++))
    # write current line to a tmp file with name $file (not needed for counting)
    echo $LINE > $file
    # this checks the return code of echo (not needed for writing; just for demo)
    if [ $? -ne 0 ]
     then echo "Error in writing to file \${file}; check its permissions!"
    fi
done

echo "Number of lines: $count"
echo "The last line of the file is: \`cat \${ file }\`"

# Note: You can achieve the same by just using the tool wc like this
echo "Expected number of lines: \`wc - l $in\`"

# restore stdin from filedescriptor 10
# and close filedescriptor 10
exec 0<&10 10<&-
`;
export const PlaintextSample = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec cursus aliquet sapien, sed rhoncus leo ullamcorper ornare. Interdum et malesuada fames ac ante ipsum primis in faucibus. Phasellus feugiat eleifend nisl, aliquet rhoncus quam scelerisque vel. Morbi eu pellentesque ex. Nam suscipit maximus leo blandit cursus. Aenean sollicitudin nisi luctus, ornare nibh viverra, laoreet ex. Donec eget nibh sit amet dolor ornare elementum. Morbi sollicitudin enim vitae risus pretium vestibulum. Ut pretium hendrerit libero, non vulputate ante volutpat et. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Nullam malesuada turpis vitae est porttitor, id tincidunt neque dignissim. Integer rhoncus vestibulum justo in iaculis. Praesent nec augue ut dui scelerisque gravida vel id velit. Donec vehicula feugiat mollis. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas.

Praesent diam lorem, luctus quis ullamcorper non, consequat quis orci. Ut vel massa vel nunc sagittis porttitor a vitae ante. Quisque euismod lobortis imperdiet. Vestibulum tincidunt vehicula posuere. Nulla facilisi. Donec sodales imperdiet risus id ullamcorper. Nulla luctus orci tortor, vitae tincidunt urna aliquet nec. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Etiam consequat dapibus massa. Sed ac pharetra magna, in imperdiet neque. Nullam nunc nisi, consequat vel nunc et, sagittis aliquam arcu. Aliquam non orci magna. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Sed id sem ut sem pulvinar rhoncus. Aenean venenatis nunc eget mi ornare, vitae maximus lacus varius. Quisque quis vestibulum justo.

Donec euismod luctus volutpat. Donec sed lacinia enim. Vivamus aliquam elit cursus, convallis diam at, volutpat turpis. Sed lacinia nisl in auctor dapibus. Nunc turpis mi, mattis ut rhoncus id, lacinia sed lectus. Donec sodales tellus quis libero gravida pretium et quis magna. Etiam ultricies mollis purus, eget consequat velit. Duis vitae nibh vitae arcu tincidunt congue. Maecenas ut velit in ipsum condimentum dictum quis eget urna. Sed mattis nulla arcu, vitae mattis ligula dictum at.

Praesent at dignissim dolor. Donec quis placerat sem. Cras vitae placerat sapien, eu sagittis ex. Mauris nec luctus risus. Cras imperdiet semper neque suscipit auctor. Mauris nisl massa, commodo sit amet dignissim id, malesuada sed ante. Praesent varius sapien eget eros vehicula porttitor.

Mauris auctor nunc in quam tempor, eget consectetur nisi rhoncus. Donec et nulla imperdiet, gravida dui at, accumsan velit. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Proin sollicitudin condimentum auctor. Sed lacinia eleifend nisi, id scelerisque leo laoreet sit amet. Morbi congue augue a malesuada pulvinar. Curabitur nec ante finibus, commodo orci vel, aliquam libero. Morbi molestie purus non nunc placerat fermentum. Pellentesque commodo ligula sed pretium aliquam. Praesent ut nibh ex. Vivamus vestibulum velit in leo suscipit, vitae pellentesque urna vulputate. Suspendisse pretium placerat ligula eu ullamcorper. Nam eleifend mi tellus, ut tristique ante ultricies vitae. Quisque venenatis dapibus tellus sit amet mattis. Donec erat arcu, elementum vel nisl at, sagittis vulputate nisi.
`;

export const YamlSample = `# some comment here
functions:
- id: greeter
  image: direktiv/greeting:v3
  type: knative-workflow
- id: solve2
  image: direktiv/solve:v3
  type: knative-workflow
description: 1
states:
- id: event-xor
  type: eventXor
  timeout: PT1H
  events:
  - event: 
      type: solveexpressioncloudevent
    transition: solve
  - event: 
      type: greetingcloudevent
    transition: greet
- id: greet
  type: action
  action:
    function: greeter
    input: jq(.greetingcloudevent.data)
  transform: 
    greeting: jq(.return.greeting)
- id: solve
  type: action
  action:
    function: solve2
    input: jq(.solveexpressioncloudevent.data)
  transform: 
    solvedexpression: jq(.return)
`;
