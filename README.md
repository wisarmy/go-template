# go-template




# Database
## Install ent command tool
```
go get entgo.io/ent/cmd/ent

```

## Create schema
```
go run -mod=mod entgo.io/ent/cmd/ent new User Role

```

## Generate ent code
```
go generate ./ent
```
