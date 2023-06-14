import { h } from "dom-chef"
import formatter from "numbuffix"
import { MetadataReponse } from "../apitypes"


export default function statistics(metadata: MetadataReponse) {
	let repos = formatter(metadata.CountOfAllRepos, '')
	let stars = formatter(metadata.CountOfAllStars, '')

    return (
        <div id="statistics" className="my-3 p-3 bg-body rounded shadow-sm">
            <h3>Statistics</h3>
            <p>Storing information about <b>{repos}</b> repositories with <b>{stars}</b> stars combined</p>
        </div>
    )
}