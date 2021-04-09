<script lang="ts">
    export let value:string
    export let options
    export let label:string
    export let id:string
</script>

<label for={id}>{label}</label>
<select bind:value {id}>
    {#each options as option}
        <option value={option.value}>{option.label}</option>
    {/each}
</select>
