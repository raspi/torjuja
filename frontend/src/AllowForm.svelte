<script lang="ts">
    import {AllowDTO} from './dto'

    async function formSubmit(evt) {
        let errorTarget = document.querySelectorAll('div.errors ul')
        errorTarget.forEach((k) => {
            k.innerHTML = ""
        })

        let dto = new AllowDTO()
        dto.fqdn = evt.target.fqdn.value

        if (dto.fqdn === '') {
            await addError('empty')
            return
        }

        const response: Response = await fetch("/api/v1/allow", {
            method: 'POST',
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(dto)
        })

        try {
            const {data, errors} = await response.json()
            if (response.ok) {
                console.log('yepa')
            } else {
                console.log(errors)
                console.log(data)
            }
        } catch (ex) {
            if (response.status == 500) {
                await addError('internal server error')
                return
            }

            await addError(ex)
        }
    }

    async function addError(errstr) {
        let target = document.querySelectorAll('div.errors ul')

        target.forEach((v, k) => {
            let li = document.createElement('li')
            li.textContent = errstr
            v.appendChild(li)
        })
    }

</script>

<h2>Allow</h2>

<div class="errors">
    <ul></ul>
</div>

<form action="#" method="post" on:submit|preventDefault={formSubmit}>
    <table>
        <tr>
            <td><label for="fqdn">Allow</label></td>
            <td><input id="fqdn" type="text" value=""></td>
        </tr>
        <tr>
            <td><label for="submit">Submit</label></td>
            <td><input id="submit" type="submit" value="Send"></td>
        </tr>
    </table>
</form>

<div class="errors">
    <ul></ul>
</div>

<style>
    div.errors {
        background: #3e0000;
        color: #eeeeee;
    }
</style>