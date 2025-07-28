const formatTimer = seconds => {
  let minutes = parseInt(seconds / 60).toString()
  seconds = parseInt(seconds % 60).toString()

  let output = minutes >= 10 ? `${minutes}` : `${minutes}`
  output += seconds >= 10 ? `:${seconds}` : `:0${seconds}`

  return output
}

export { formatTimer }
