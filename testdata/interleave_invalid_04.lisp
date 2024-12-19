;;error:4:1-25:source column Y has incompatible length multiplier
(defcolumns X Y)
(definterleaved A (X Y))
(definterleaved B (A Y))
