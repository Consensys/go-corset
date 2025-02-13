(defcolumns (A :binary@loob) (B :binary@loob) (C :@loob))
(defconstraint c1 () (if (- A B) C))
(defconstraint c2 () (if (- B A) C))
