<script lang="ts">
    import {AllowDTO} from './dto'
    import Form from "./form/Form.svelte";

    async function formSubmit(evt) {
        let errorTarget = document.querySelectorAll('div.errors ul')
        errorTarget.forEach((k) => {
            k.innerHTML = ""
        })

        let dto = new AllowDTO()
        dto.fqdn = evt.fqdn

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

    //   Our field representation, let's us easily specify several inputs
    let fields = [
        {
            name: "fqdn",
            type: "Input",
            value: "",
            placeholder: "FQDN...",
            label: "FQDN",
        }
    ]

</script>

<h2>Allow</h2>

<div class="errors">
    <ul></ul>
</div>

<Form onSubmit={formSubmit} {fields}/>

<div class="errors">
    <ul></ul>
</div>

<style>
    div.errors {
        background: #3e0000;
        color: #eeeeee;
    }
</style>