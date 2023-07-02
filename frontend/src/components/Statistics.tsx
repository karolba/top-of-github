import { h } from "dom-chef"
import formatter from "numbuffix"
import { MetadataReponse } from "../apitypes"


export default function Statistics(metadata: MetadataReponse): JSX.Element {
    let repos = formatter(metadata.CountOfAllRepos, '')
    let stars = formatter(metadata.CountOfAllStars, '')

    return (
        <div id="statistics" className="my-3 p-3 bg-body rounded shadow-sm">
            <h5>Statistics</h5>
            <p>Storing information about <b>{repos}</b> repositories with <b>{stars}</b> stars combined - collected every GitHub repository with at least 5 stars.</p>
        </div>
    )
}