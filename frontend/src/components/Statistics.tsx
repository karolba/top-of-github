import { h } from "dom-chef"
import { MetadataReponse } from "../apitypes"
import { humanizeNumber } from "../humanizeNumber"


export default function Statistics(metadata: MetadataReponse): JSX.Element {
    let preciseRepos = metadata.CountOfAllRepos.toLocaleString('en')
    let approxRepos = humanizeNumber(metadata.CountOfAllRepos)

    let preciseStars = metadata.CountOfAllStars.toLocaleString('en')
    let approxStars = humanizeNumber(metadata.CountOfAllStars)

    return (
        <div id="statistics" className="my-3 p-3 bg-body rounded shadow-sm">
            <h5>Statistics</h5>
            <p>
                Storing information about
                { } <span title={preciseRepos}><b>{approxRepos}</b></span>
                { } repositories with
                { } <span title={preciseStars}><b>{approxStars}</b></span>
                { } stars combined - collected every GitHub repository with at least <b>5</b> stars.</p>
        </div>
    )
}