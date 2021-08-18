# go-biggier
a go lang big integer lib

**refer to java BigInteger**

> In the go language, we know that there are only `byte`, `int32`, and `int64` for integers, while float have only `float32` and `float64`. Sometimes we need to use large number operations according to your needs. The `big` package in go with `NewInt` and `NewFloat`, but there are fewer APIs, and only basic types can be used for initialization. You cannot use a string for initialization. So I developed `go-bigger` with reference to Java's large number classes and provided a rich API calls.

## 1. Import module
> go get github.com/sineycoder/go-bigger

## 2. BigInteger

**In BigInteger, we cached |x| < 16 BigInteger**

> you can use `bigger.NewBigIntegerString("123123")` or `bigger.BigIntegerValueOf(6782613786431)` to initialize a BigInteger. If use `BigIntegerValueOf` and whithin 16, it returns a chache BigInteger.

### 2.1 Add

```
func main() {
	a := bigger.BigIntegerValueOf(6782613786431)
	b := bigger.BigIntegerValueOf(-678261378231)
	res := a.Add(b)
	fmt.Println(res.String())
}

// result：6104352408200
```

### 2.2 Subtract

```
func main() {
	a := bigger.BigIntegerValueOf(6782613786431)
	b := bigger.BigIntegerValueOf(-678261378231)
	res := a.Subtract(b)
	fmt.Println(res.String())
}

// result：7460875164662
```

### 2.3 Divide

```
func main() {
	a := bigger.BigIntegerValueOf(6782613786431)
	b := bigger.BigIntegerValueOf(-678261378231)
	res := a.Divide(b)
	fmt.Println(res.String())
}

// result： -10
```

### 2.4 Multiply

```
func main() {
	a := bigger.BigIntegerValueOf(6782613786431)
	b := bigger.BigIntegerValueOf(-678261378231)
	res := a.Multiply(b)
	fmt.Println(res.String())
}

// result：-4600384974793271546583561
```

## 3.BigDecimal

> you can use `bigger.NewBigDecimalString("123123.111")` or `bigger.BigDecimalValueOf(6782613786431.111)` to initialize a BigInteger. If use `BigDecimalValueOf` and whithin 10, it returns a chache BigDecimal.

### 3.1 Add

```
func main() {
	a := bigger.BigDecimalValueOf(6782613786431)
	b := bigger.BigDecimalValueOf(-678261378231)
	res := a.Add(b)
	fmt.Println(res.String())
}

// result：6104352408200
```

### 3.2 Subtract

```
func main() {
	a := bigger.BigDecimalValueOf(6782613786431)
	b := bigger.BigDecimalValueOf(-678261378231)
	res := a.Subtract(b)
	fmt.Println(res.String())
}

// result：7460875164662
```

### 3.3 Divide

```
func main() {
	a := bigger.BigDecimalValueOf(6782613786431)
	b := bigger.BigDecimalValueOf(-678261378231)
	res := a.Divide(b, 12, bigger.ROUND_HALF_UP)
	fmt.Println(res.String())
}

// result： -10.000000006076 
```

### 3.4 Multiply

```
func main() {
	a := bigger.BigDecimalValueOf(6782613786431)
	b := bigger.BigDecimalValueOf(-678261378231)
	res := a.Multiply(b)
	fmt.Println(res.String())
}

// result：-4600384974793271546583561
```

if you want to set a precision, you can use `setScale()`. Remenber: The return value must be assigned. (e.g. `res = res.setScale(12, bigger.ROUND_HALF_UP)`)
