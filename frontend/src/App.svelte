<script lang="ts">

    const blockedEventsURL = '/events/blocked'
	let sseEvents:EventSource = new EventSource(blockedEventsURL)
    sseEvents.onmessage = evt => {
        let eventList = document.querySelector('#events')

        let newRow = document.createElement('tr')
        let newCell = document.createElement('td')
        newCell.textContent = evt.data
        newRow.appendChild(newCell)

        eventList.prepend(newRow)
    }

    sseEvents.onerror = evt => {
        console.log(evt)
    }

</script>

<main>
    <table>
        <thead>
        <tr>
            <th>Req</th>
        </tr>
        </thead>
        <tbody id="events">
        </tbody>
    </table>
</main>

<style>
</style>