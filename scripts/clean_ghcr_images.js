(async function ()   {
    let githubToken = "some secrete"

    for(;;) {
        let list = await fetch('https://api.github.com/orgs/direktiv/packages/container/direktiv/versions', {
            headers: {
                "Authorization": `Bearer ${githubToken}`,
            },
        });
        if(list.status !== 200 ) {
            throw Error(`fetching images list with status: ${list.status}`)
        }
        list = await list.json()

        if(list.length < 2) {
            return
        }

        for(let i = 0;i< list.length; i++) {
            let res = await fetch(`https://api.github.com/orgs/direktiv/packages/container/direktiv/versions/${list[i].id}`, {
                method: "DELETE",
                headers: {
                    "Authorization": `Bearer ${githubToken}`,
                },
            });
            console.log(`deleting image: ${list[i].id}`);
        }
    }
})()
