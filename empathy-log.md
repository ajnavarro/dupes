# Empathy log

- Go to bblfsh getting started
- Start docker container
- Update drivers
    - _There is no an easy way to update all drivers_
        - That command should appear on documentation: `docker exec -it bblfshd bblfshctl driver install --recommended --update`
- Next documentation point: UAST querying
    - _I don't need that at that point, I need a way to connect to bblfsh programatically_
- Next documentation point: Language clients
    - _a little bit strange name for bblfsh clients_
- Go to bblfsh/client-go github repository
- Copy the example from readme to test that it's working
- All working, go back to UAST nodes documentation
- I have the necessity of know the language of the file before hand to use bblfsh
    - _talk about enry on bblfsh documentation to fill that necessity_
- Started to play a bit with Xpath
    - _There is no place on UAST documentation where I can check which roles I should use to get functions_
- Using node.Hash() method, that after 20 min testing is always returning 0
    - _methods that are not implemented should panic or return an error_
```
// Hash returns the hash of the node.
func (n *Node) Hash() Hash {
	return n.HashWith(IncludeChildren)
}

// HashWith returns the hash of the node, computed with the given set of fields.
func (n *Node) HashWith(includes IncludeFlag) Hash {
	//TODO
	return 0
}
```

- Trying to get positions from UAST. It's not clear when positions will be returned