<script lang="ts">
    import {AllowDTO} from './dto'

    async function formSubmit(evt) {
        let dto = new AllowDTO()
        dto.fqdn = evt.target.fqdn.value

        const response = await fetch("/api/v1/allow", {
            method: 'POST',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(dto)
        });

        const data = await response.json();
        console.log(data)
    }

    const blockedEventsURL = '/events/blocked'
    let sseEvents: EventSource = new EventSource(blockedEventsURL)
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
    <form action="#" method="post" on:submit|preventDefault={formSubmit}>
        <label for="fqdn">Allow</label>
        <input id="fqdn" type="text" value="">
        <label for="submit">Submit</label>
        <input id="submit" type="submit" value="Send">
    </form>

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