;;error:15:53-54:unknown symbol
;;error:16:53-54:unknown symbol
(defpurefun ((vanishes! :@loob :force) e0) e0)
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
(defconstraint c1 (:perspective p1) (vanishes! (- A C)))
(defconstraint c2 (:perspective p2) (vanishes! (- A B)))
