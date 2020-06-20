
/**

  # Levenshtein

  Computes Levenshtein distance between two strings.

*/

package levenshtein

import (
  "fmt"
)


func Distance(s, t string) int {

  d := make([][]int, len(s)+1)

  for i := range d {
    d[i] = make([]int, len(t)+1)
  }

  for i := range d {
    d[i][0] = i
  }

  for j := range d[0] {
    d[0][j] = j
  }

  for j := 1; j <= len(t); j++ {

    for i := 1; i <= len(s); i++ {

      if s[i-1] == t[j-1] {

        d[i][j] = d[i-1][j-1]

      } else {

        min := d[i-1][j]

        if d[i][j-1] < min {
          min = d[i][j-1]
        }

        if d[i-1][j-1] < min {
          min = d[i-1][j-1]
        }

        d[i][j] = min + 1

      }

    }

  }

  return d[len(s)][len(t)]
}

/**
 * Returns closest the closest match to the input from a list of options. Options should be in order of preference for
 * resolving ties.
 */
func Suggest(input string, options []string) string {
  var lowestDistance int
  var lowestIndex int

  for i := range options {

    distance := Distance(input, options[i])

    if i == 0 || distance < lowestDistance {
      lowestDistance = distance
      lowestIndex = i
    }
  }
  return options[lowestIndex]
}


func Test() {

  fmt.Println(Distance("git push", "git push") == 0)
  fmt.Println(Distance("commit", "commir") == 1)
  fmt.Println(Distance("kitten", "sitting") == 3)

  fmt.Println(Suggest("gut", []string{"git", "vim", "curl"}) == "git")
  fmt.Println(Suggest("beap", []string{"start", "stop", "beat"}) == "beat")

}