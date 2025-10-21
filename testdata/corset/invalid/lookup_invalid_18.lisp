;;error:7:23-24:conflicting context
(module m1)
(defcolumns (X :i64))

(module m2)
(defcolumns (A :i1) (B :i64))
(defclookup m1 (m2.B) A (m1.X))
