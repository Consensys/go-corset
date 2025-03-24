;;error:15:28-29:unknown symbol
;;error:16:28-29:unknown symbol
;;
;;
(defcolumns
    ;; Column (not in perspective)
    (A :i16)
    ;; Selector column for perspective p1
    (P :binary@prove)
    ;; Selector column for perspective p2
    (Q :binary@prove))

(defperspective p1 P ((B :byte)))
(defperspective p2 Q ((C :byte)))
(defconstraint c1 () (== A B))
(defconstraint c2 () (== A C))
