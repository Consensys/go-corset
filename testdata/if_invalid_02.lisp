;;error:3:22-34:incorrect number of arguments
(defcolumns (A :i16) (B :i16) (C :i16) (D :i16))
(defconstraint c1 () (if A B C D))
