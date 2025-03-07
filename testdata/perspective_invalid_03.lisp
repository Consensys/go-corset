;;error:15:38-39:unknown symbol
;;error:16:38-39:unknown symbol
(defpurefun ((vanishes! :𝔽@loob :force) x) x)
;;
(defcolumns
    ;; Column (not in perspective)
    A
    ;; Selector column for perspective p1
    (P :binary@prove)
    ;; Selector column for perspective p2
    (Q :binary@prove))

(defperspective p1 P ((B :byte)))
(defperspective p2 Q ((C :byte)))
(defconstraint c1 () (vanishes! (- A B)))
(defconstraint c2 () (vanishes! (- A C)))
