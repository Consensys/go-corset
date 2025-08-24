;;error:3:15-29:incompatible type (u32)
(defcolumns (X :i32))
(defcomputed ((Y :i16@prove)) (id X))
