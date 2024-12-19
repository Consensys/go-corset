;;error:4:1-25:source column A has incompatible length multiplier
(defcolumns X Y)
(definterleaved A (X Y))
(definterleaved B (X A))
