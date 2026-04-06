package dental

import (
  "fmt"
  "strings"
  "testing"
)

func brack(txt string) string {
  lines := strings.Split(txt, "\n")
  for i, line := range lines {
    lines[i] = fmt.Sprintf("[%v]", line)
  }
  return strings.Join(lines, "\n")
}

func TestTestingInfra(t *testing.T) {
  if brack("hi") != "[hi]" {
    t.Error("[hi]")
  }
  if brack("hi\n  there") != "[hi]\n[  there]" {
    t.Error(":multiline")
  }
}

func ExampleDental_LineLeadingSpaces() {
  var d *Dental = &Dental{}
  fmt.Println(d.LineLeadingSpaces("   hi"))
  // Output: 3
}

func ExampleDental_MinLeadingSpaces() {
  var d *Dental = &Dental{}
  fmt.Println(d.MinLeadingSpaces([]string{
    "   hi",
    " there",
  }))
  // Output: 1
}

func ExampleSpaces() {
  fmt.Println(brack(Spaces(3)))
  // Output: [   ]
}

func ExampleDental_IndentationLevel() {
  var d *Dental = New(4)
  indent, remainder := d.IndentationLevel([]string{
    "     // this is indented 5 spaces",
    "     function hello() {",
    "         // this is indented 9 spaces",
    "         console.log('hi!');",
    "     }",
  })
  fmt.Printf("indent=%v, remainder=%v\n", indent, remainder)
  // Output: indent=1, remainder=1
}

func ExampleDental_DedentLines() {
  lines := []string{
    "       the mouse ran up the clock",
    "    the boat sailed into the dock",
    "                the mouse had fun",
    "              the journey's begun",
    "       the mouse found a wet sock",
  }
  New(2).DedentLines(lines)
  for _, line := range lines {
    fmt.Println(brack(line))
  }
  // Output:
  // [   the mouse ran up the clock]
  // [the boat sailed into the dock]
  // [            the mouse had fun]
  // [          the journey's begun]
  // [   the mouse found a wet sock]
}

func ExampleDental_DedentBlock() {
  fmt.Println(brack(New(2).DedentBlock(`
    Once upon a time a girl spent a dime
       on a little ring of sorts hailed from distant ports
         if I had a nickel and another sickle
         winds a little fickle with that itchin' tickle
      I'd find a ring give it a fling
      sail past those birds who chirp and sing
    alas all this was not to be
    the ring for me was lost at sea
  `)))
  // Output:
  // []
  // [Once upon a time a girl spent a dime]
  // [   on a little ring of sorts hailed from distant ports]
  // [     if I had a nickel and another sickle]
  // [     winds a little fickle with that itchin' tickle]
  // [  I'd find a ring give it a fling]
  // [  sail past those birds who chirp and sing]
  // [alas all this was not to be]
  // [the ring for me was lost at sea]
  // []
}

func ExampleDental_IndentBlock() {
  block := `
    Once upon a time a girl spent a dime
      on a little ring of sorts hailed from distant ports
        if I had a nickel and another sickle
        winds a little fickle with that itchin' tickle
      I'd find a ring give it a fling
      sail past those birds who chirp and sing
    alas all this was not to be
    the ring for me was lost at sea
  `
  d := New(2)
  fmt.Println("indenting block by 3 (6 spaces)")
  indented := d.IndentBlock(block, 3) 
  fmt.Println(brack(indented)) 
  fmt.Printf("leading spaces is %v\n", d.MinLeadingSpaces([]string{
    indented,
  }))
  indent, remainder := d.IndentationLevel([]string{ indented })
  fmt.Printf("indentation level is %v with %v remainder\n", indent, remainder)
  fmt.Println("---")

  fmt.Println("indenting block by -1 (-2 spaces)")
  indented = d.IndentBlock(block, -1) 
  fmt.Println(brack(indented)) 
  fmt.Printf("leading spaces is %v\n", d.MinLeadingSpaces([]string{
    indented,
  }))
  indent, remainder = d.IndentationLevel([]string{ indented })
  fmt.Printf("indentation level is %v with %v remainder\n", indent, remainder)
  // Output:
  // indenting block by 3 (6 spaces)
  // []
  // [          Once upon a time a girl spent a dime]
  // [            on a little ring of sorts hailed from distant ports]
  // [              if I had a nickel and another sickle]
  // [              winds a little fickle with that itchin' tickle]
  // [            I'd find a ring give it a fling]
  // [            sail past those birds who chirp and sing]
  // [          alas all this was not to be]
  // [          the ring for me was lost at sea]
  // []
  // leading spaces is 10
  // indentation level is 5 with 0 remainder
  // ---
  // indenting block by -1 (-2 spaces)
  // []
  // [  Once upon a time a girl spent a dime]
  // [    on a little ring of sorts hailed from distant ports]
  // [      if I had a nickel and another sickle]
  // [      winds a little fickle with that itchin' tickle]
  // [    I'd find a ring give it a fling]
  // [    sail past those birds who chirp and sing]
  // [  alas all this was not to be]
  // [  the ring for me was lost at sea]
  // []
  // leading spaces is 2
  // indentation level is 1 with 0 remainder
}

func ExampleDental_SetBlockIndentation() {
  d := New(2)
  indented := d.SetBlockIndentation(`
    Once upon a time a girl spent a dime
      on a little ring of sorts hailed from distant ports
        if I had a nickel and another sickle
        winds a little fickle with that itchin' tickle
      I'd find a ring give it a fling
      sail past those birds who chirp and sing
    alas all this was not to be
    the ring for me was lost at sea
  `, 3)
  fmt.Println(brack(indented)) 
  fmt.Printf("leading spaces is %v\n", d.MinLeadingSpaces([]string{
    indented,
  }))
  indent, remainder := d.IndentationLevel([]string{ indented })
  fmt.Printf("indentation level is %v with %v remainder\n", indent, remainder)
  // Output:
  // []
  // [      Once upon a time a girl spent a dime]
  // [        on a little ring of sorts hailed from distant ports]
  // [          if I had a nickel and another sickle]
  // [          winds a little fickle with that itchin' tickle]
  // [        I'd find a ring give it a fling]
  // [        sail past those birds who chirp and sing]
  // [      alas all this was not to be]
  // [      the ring for me was lost at sea]
  // []
  // leading spaces is 6
  // indentation level is 3 with 0 remainder
}

