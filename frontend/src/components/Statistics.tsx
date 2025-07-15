import { h } from "dom-chef"
import { MetadataReponse } from "../apitypes"
import { humanizeNumber } from "../humanizeNumber"

function timeAgo(date: Date) {
    const formatter = new Intl.RelativeTimeFormat('en', {numeric: 'always'});
    const secondsSince = (date.getTime() - Date.now()) / 1000;

    const SECONDS_IN_MINUTE = 60
    const SECONDS_IN_HOUR = 60 * 60
    const SECONDS_IN_DAY = 60 * 60 * 24
    const SECONDS_IN_WEEK = 60 * 60 * 24 * 7

    if (Math.abs(secondsSince) > SECONDS_IN_DAY * 2)
        return formatter.format(Math.trunc(secondsSince / SECONDS_IN_DAY), 'days')

    if (Math.abs(secondsSince) > SECONDS_IN_HOUR)
        return formatter.format(Math.trunc(secondsSince / SECONDS_IN_HOUR), 'hours')

    if (Math.abs(secondsSince) > SECONDS_IN_MINUTE)
        return formatter.format(Math.trunc(secondsSince / SECONDS_IN_MINUTE), 'minutes')

    return formatter.format(Math.trunc(secondsSince), 'seconds')
}

export default function Statistics(metadata: MetadataReponse): JSX.Element {
    let preciseRepos = metadata.CountOfAllRepos.toLocaleString('en')
    let approxRepos = humanizeNumber(metadata.CountOfAllRepos)

    let preciseStars = metadata.CountOfAllStars.toLocaleString('en')
    let approxStars = humanizeNumber(metadata.CountOfAllStars)

    let preciseLastSyncTime = metadata.LastSyncTime
    let approxLastSyncTime = timeAgo(new Date(preciseLastSyncTime))

    return (
        <div id="statistics" className="my-3 p-3 bg-body rounded shadow-sm">
            <h5>Statistics</h5>
            <p>
                Storing information on
                { } <span title={preciseRepos}><b>{approxRepos}</b></span>
                { } repositories with
                { } <span title={preciseStars}><b>{approxStars}</b></span>
                { } stars combined - collected every GitHub repository with at least <b>5</b> stars.
            </p>
            <p>
                Last sync with GitHub finished
                { } <span title={preciseLastSyncTime}><b>{approxLastSyncTime}</b></span>.
            </p>
        </div>
    )
}
