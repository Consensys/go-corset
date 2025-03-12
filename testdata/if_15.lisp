(defcolumns (A :binary@loob) (B :binary@loob) (C :i16@loob))
(defconstraint c1 () (if (- A B) C))
(defconstraint c2 () (if (- B A) C))
