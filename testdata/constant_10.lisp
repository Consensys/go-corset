(defcolumns (X :@loob) Y)
(defconst
  N 4
  TWO_N (^ 2 N))

;; X == Y * 2^n
(defconstraint c1 () (- X (* Y TWO_N)))
