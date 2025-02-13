(defcolumns (A :binary@loob) (B :@loob) (C :@loob))
(defconstraint c1 () (if (- A 1) C B))
(defconstraint c2 () (if (- 1 A) C B))
